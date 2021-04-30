package pauper

import (
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

// SkipWS returns an index to the first byte that's not whitespace in
// the sense of the JSON spec, or the length of the buffer.
func SkipWS(buf []byte) int {
	for i, c := range buf {
		switch c {
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
func GetString(buf []byte) ([]byte, int, error) {
	skip := SkipWS(buf)
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

func GetNumber(buf []byte) (float64, int, error) {
	start := SkipWS(buf)
	i := start
	if len(buf) <= i {
		return 0, 0, ErrNoNumberHere
	}
	// sign
	if buf[i] == '-' {
		i++
	}
	if len(buf) <= i {
		return 0, 0, ErrNoNumberHere
	}
	// integer
	switch buf[i] {
	case '0':
		i++
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		i++
		for len(buf) > i && buf[i] >= '0' && buf[i] <= '9' {
			i++
		}
	default:
		return 0, 0, ErrNoNumberHere
	}
	if len(buf) > i && buf[i] == '.' {
		// fraction
		i++
		for len(buf) > i && buf[i] >= '0' && buf[i] <= '9' {
			i++
		}
	}
	if len(buf) > i && buf[i]|32 == 'e' {
		// exponent
		i++
		if len(buf) > i && (buf[i] == '+' || buf[i] == '-') {
			i++
		}
		for len(buf) > i && buf[i] >= '0' && buf[i] <= '9' {
			i++
		}
	}
	f, err := strconv.ParseFloat(string(buf[start:i]), 64)
	return f, i, err
}
