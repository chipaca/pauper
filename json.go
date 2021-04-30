package pauper

import (
	"unicode/utf16"
	"unicode/utf8"
)

// SkipWS returns an index to the first byte that's not whitespace in
// the sense of the JSON spec, or the length of the buffer.
func SkipWS(buf []byte, skip int) int {
	for i := skip; i < len(buf); i++ {
		switch buf[i] {
		default:
			return i
		case ' ', '\n', '\r', '\t':
		}
	}
	return len(buf)
}

const (
	ErrNoStringHere   = Error("no string here")
	ErrNoNumberHere   = Error("no number here")
	ErrNotImplemented = Error("not implemented")
)

// GetString returns the initial string (as a []byte), and a skip to
// just after the closing '"'. `buf` should start with any amount of
// whitespace and then an opening '"'. The passed-in buffer may be
// modified in place, and the returned buffer may reference it.
func GetString(buf []byte, skip int) ([]byte, int, error) {
	skip = SkipWS(buf, skip)
	if len(buf) < skip+2 || buf[skip] != '"' {
		return nil, 0, ErrNoStringHere
	}
	skip++
	start := skip
	end := len(buf) - 1
	hasEscapes := false
	for ; skip < end && buf[skip] != '"'; skip++ {
		if buf[skip] == '\\' {
			skip++
			hasEscapes = true
		}
	}
	if skip > end || buf[skip] != '"' {
		return nil, 0, ErrNoStringHere
	}
	if !hasEscapes {
		return buf[start:skip], skip + 1, nil
	}
	raw := buf[:0]
	for i := start; i < skip; {
		r, w := utf8.DecodeRune(buf[i:])
		if r == utf8.RuneError {
			return nil, 0, ErrNoStringHere
		}
		i += w
		switch r {
		case '\\':
			r, w := utf8.DecodeRune(buf[i:])
			if r == utf8.RuneError {
				return nil, 0, ErrNoStringHere
			}
			i += w
			switch r {
			case '"', '/', '\\':
				raw = append(raw, byte(r))
			case 'b':
				raw = append(raw, '\b')
			case 'f':
				raw = append(raw, '\f')
			case 'n':
				raw = append(raw, '\n')
			case 'r':
				raw = append(raw, '\r')
			case 't':
				raw = append(raw, '\t')
			case 'u':
				u := u4(buf, i)
				if u < 0 {
					return nil, 0, ErrNoStringHere
				}
				i += 4
				if utf16.IsSurrogate(u) {
					if len(buf) < i+6 || buf[i] != '\\' || buf[i+1] != 'u' {
						return nil, 0, ErrNoStringHere
					}
					i += 2
					v := u4(buf, i)
					if v < 0 {
						return nil, 0, ErrNoStringHere
					}
					i += 4
					u = utf16.DecodeRune(u, v)
				}
				raw = append(raw, string(u)...)
			}
		default:
			raw = append(raw, string(r)...)
		}
	}
	return raw, skip + 1, nil
}

func u4(buf []byte, start int) rune {
	var r rune
	for i := start; i < start+4; i++ {
		var v byte
		c := buf[i]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			v = c - '0'
		case 'a', 'b', 'c', 'd', 'e', 'f':
			v = c - 'a' + 10
		case 'A', 'B', 'C', 'D', 'E', 'F':
			v = c - 'A' + 10
		default:
			return -1
		}

		r <<= 4
		r |= rune(v)
	}
	return r
}

// GetInt parses a subset of the JSON 'number', namely the 'integer'
// part of it. If there's a decimal point or an exponent it will
// fail. If there are too many digits for an int64 it'll fail.
func GetInt(buf []byte, skip int) (int64, int, error) {
	start := SkipWS(buf, skip)
	i := start
	if len(buf) <= i {
		return 0, 0, ErrNoNumberHere
	}
	var neg bool
	if buf[i] == '-' {
		i++
		if len(buf) <= i {
			return 0, 0, ErrNoNumberHere
		}
		neg = true
	}
	var n int64
	for m := i + 20; i < m && len(buf) > i && '0' <= buf[i] && buf[i] <= '9'; i++ {
		n *= 10
		n += int64(buf[i] - '0')
	}
	if neg {
		n = -n
	}
	if len(buf) <= i {
		return n, i, nil
	}
	// if the number is finished, the next byte will be
	// whitespace, or a comma, or a closing bracket or brace
	switch buf[i] {
	case ' ', '\n', '\r', '\t', ',', ']', '}':
		return n, i, nil
	}
	return n, i, ErrNoNumberHere
}

/*

var float64pow10 = []float64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
	1e20, 1e21, 1e22,
}

func GetFloat(buf []byte) (float64, int, error) {
	start := SkipWS(buf)
	i := start
	if len(buf) <= i {
		return 0, 0, ErrNoNumberHere
	}

	// simplified from strconv.ParseFloat

	// sign
	var neg bool
	if buf[i] == '-' {
		i++
		neg = true
	}
	if len(buf) <= i {
		return 0, 0, ErrNoNumberHere
	}
	// mantissa
	const maxMantDigits = 19
	nd := 0     // the number of digits seen
	ndMant := 0 // the number in the mantissa
	dp := 0     // the position of the '.'
	var mantissa uint64
	sawdot := false
	sawdig := false
	for ; len(buf) > i; i++ {
		switch c := buf[i]; true {
		case c == '.':
			if sawdot {
				return 0, 0, ErrNoNumberHere
			}
			sawdot = true
			dp = nd
			continue
		case '0' <= c && c <= '9':
			sawdig = true
			if c == '0' && nd == 0 {
				dp--
				continue
			}
			nd++
			if ndMant < maxMantDigits {
				mantissa *= 10
				mantissa += uint64(c - '0')
				ndMant++
			}
			continue
		}
		break
	}
	if !sawdig {
		return 0, 0, ErrNoNumberHere
	}
	if !sawdot {
		dp = nd
	}
	if len(buf) > i && (buf[i] == 'e' || buf[i] == 'E') {
		i++
		if len(buf) <= i {
			return 0, 0, ErrNoNumberHere
		}
		esign := 1
		switch buf[i] {
		case '-':
			esign = -1
			fallthrough
		case '+':
			i++
		}
		if i >= len(buf) || buf[i] < '0' || buf[i] > '9' {
			return 0, 0, ErrNoNumberHere
		}
		e := 0
		for ; i < len(buf) && '0' <= buf[i] && buf[i] <= '9'; i++ {
			if e < 10000 {
				e = e*10 + int(buf[i]-'0')
			}
		}
		dp += e * esign
	}

	var exp int
	if mantissa != 0 {
		exp = dp - ndMant
	}

	f := float64(mantissa)
	if neg {
		f = -f
	}
	var float64pow10 = []float64{
		1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
		1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
		1e20, 1e21, 1e22,
	}

	switch {
	case exp > 0:
		for exp > 22 {
			exp -= 22
			f *= float64pow10[22]
		}
		f *= float64pow10[exp]
	case exp < 0:
		for exp < -22 {
			exp += 22
			f /= float64pow10[22]
		}
		f /= float64pow10[-exp]
	}
	return f, i, nil
}

*/
