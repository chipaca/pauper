package pauper_test

import (
	"testing"

	"chipaca.com/pauper"
)

const (
	allWS = " \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t" +
		" \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t \n\r\t"
)

func TestSkipWS(t *testing.T) {
	for from, expected := range map[string]int{
		allWS:           80,
		"hello":         0,
		"  two":         2,
		allWS + "  two": 82,
	} {
		got := pauper.SkipWS([]byte(from))
		if got != expected {
			t.Errorf("got %d expecting %d", got, expected)
		}
	}
}

func BenchmarkSkipWS(b *testing.B) {
	b1 := []byte(allWS)
	b2 := []byte(allWS + "hello")
	for i := 0; i < b.N; i++ {
		pauper.SkipWS(b1)
		pauper.SkipWS(b2)
	}
}

func TestGetString(t *testing.T) {
	for from, expected := range map[string]string{
		`"hello"`:           "hello",
		`   "hi"`:           "hi",
		` "hi\""`:           "hi\"",
		`     ""`:           "",
		`"árbol"`:           "árbol",
		`"\u0020"`:          " ",
		`"\uD83D\uDCA9 hi"`: "💩 hi",
	} {
		t.Run(from, func(t *testing.T) {
			got, i, err := pauper.GetString([]byte(from))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != expected {
				t.Errorf("GetString(%q) returned %q instead of %q", from, got, expected)
			}
			if from[i-1] != '"' {
				t.Error("(n-1) was not \"")
			}
		})
	}
	for _, from := range []string{
		``,
		`"`,
		`"\"`,
		`"hi`,
		`   "`,
	} {
		t.Run(from, func(t *testing.T) {
			_, _, err := pauper.GetString([]byte(from))
			if err != pauper.ErrNoStringHere {
				t.Errorf("GetString(%q) returned %v instead of %q", from, err, pauper.ErrNoStringHere)
			}
		})
	}
}

func BenchmarkGetString_noEscapes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pauper.GetString([]byte(`"1234567890qwertyuiopasdfghjklzxcvbnm áéíóú"`))
	}
}
func BenchmarkGetString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pauper.GetString([]byte(`"1234567890qwertyuiopasdfghjklzxcvbnm áéíóú \uD83D\uDCA9 \"💩\""`))
	}
}

func TestGetInt(t *testing.T) {
	for from, expected := range map[string]int64{
		"0":                    0,
		"1":                    1,
		"1000":                 1000,
		"9223372036854775807":  1<<63 - 1,
		"-0":                   0,
		"-1":                   -1,
		"-1000":                -1000,
		"-9223372036854775807": -(1<<63 - 1),
		"-9223372036854775808": -1 << 63,
	} {
		for _, last := range []string{"", ",", "]", "}", " "} {
			for _, start := range []string{"", " ", " \n\t\r"} {
				in := start + from + last
				t.Run(in, func(t *testing.T) {
					f, i, e := pauper.GetInt([]byte(in))
					if e != nil {
						t.Fatal(e)
					}
					if f != expected {
						t.Errorf("expected %d, got %d", expected, f)
					}
					if last != "" {
						if in[i] != last[0] {
							t.Errorf("expected i to point to the ending %s in %q, got %d (%c)", last, in, i, from[i])
						}
					}
				})
			}
		}
	}
}

func BenchmarkGetInt(b *testing.B) {
	buf := []byte("   -9223372036854775808   ")
	for i := 0; i < b.N; i++ {
		pauper.GetInt(buf)
	}
}

// func TestGetFloat(t *testing.T) {
// 	for from, expected := range map[string]float64{
// 		"0,":   0,
// 		" -0,": 0,
// 		"10,":  10,
// 		"12345678900000000000000000000000000000000000000000000000000000000000000000000000000000000000,": 1.23456789e91,
// 		"1.23,":    1.23,
// 		"1.23e4,":  12300,
// 		"1.23e+4,": 12300,
// 		"123e4,":   1230000,
// 		"123e-2,":  1.23,
// 	} {
// 		t.Run(from, func(t *testing.T) {
// 			buf := []byte(from)
// 			got, i, err := pauper.GetFloat(buf)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			if buf[i] != ',' {
// 				t.Errorf("%q at %d?", buf, i)
// 			}
// 			if got != expected {
// 				t.Errorf("%g != %g", got, expected)
// 			}
// 		})
// 	}
// }

// type atofTest struct {
// 	in  string
// 	out float64
// }

// var ErrSyntax = pauper.Error("x")
// var ErrRange = pauper.Error("y")

// // from strconv's atof tests, pruned
// var atofTests = []atofTest{
// 	{"1", 1},
// 	{"1e23", 1e+23},
// 	{"1E23", 1e+23},
// 	{"100000000000000000000000", 1e+23},
// 	{"1e-100", 1e-100},
// 	{"123456700", 1.234567e+08},
// 	{"99999999999999974834176", 9.999999999999997e+22},
// 	{"100000000000000000000001", 1.0000000000000001e+23},
// 	{"100000000000000008388608", 1.0000000000000001e+23},
// 	{"100000000000000016777215", 1.0000000000000001e+23},
// 	{"100000000000000016777216", 1.0000000000000003e+23},
// 	// {"-1", "-1"},
// 	// {"-0.1", "-0.1"},
// 	// {"-0", "-0"},
// 	// {"1e-20", "1e-20"},
// 	// {"625e-3", "0.625"},

// 	// // zeros
// 	// {"0", "0"},
// 	// {"0e0", "0"},
// 	// {"-0e0", "-0"},
// 	// {"+0e0", "0"},
// 	// {"0e-0", "0"},
// 	// {"-0e-0", "-0"},
// 	// {"+0e-0", "0"},
// 	// {"0e+0", "0"},
// 	// {"-0e+0", "-0"},
// 	// {"+0e+0", "0"},
// 	// {"0e+01234567890123456789", "0"},
// 	// {"0.00e-01234567890123456789", "0"},
// 	// {"-0e+01234567890123456789", "-0"},
// 	// {"-0.00e-01234567890123456789", "-0"},

// 	// {"0e291", "0"}, // issue 15364
// 	// {"0e292", "0"}, // issue 15364
// 	// {"0e347", "0"}, // issue 15364
// 	// {"0e348", "0"}, // issue 15364
// 	// {"-0e291", "-0"},
// 	// {"-0e292", "-0"},
// 	// {"-0e347", "-0"},
// 	// {"-0e348", "-0"},

// 	// // largest float64
// 	// {"1.7976931348623157e308", "1.7976931348623157e+308"},
// 	// {"-1.7976931348623157e308", "-1.7976931348623157e+308"},

// 	// // the border is ...158079
// 	// // borderline - okay
// 	// {"1.7976931348623158e308", "1.7976931348623157e+308"},
// 	// {"-1.7976931348623158e308", "-1.7976931348623157e+308"},

// 	// // a little too large
// 	// {"1e308", "1e+308"},

// 	// // denormalized
// 	// {"1e-305", "1e-305"},
// 	// {"1e-306", "1e-306"},
// 	// {"1e-307", "1e-307"},
// 	// {"1e-308", "1e-308"},
// 	// {"1e-309", "1e-309"},
// 	// {"1e-310", "1e-310"},
// 	// {"1e-322", "1e-322"},
// 	// // smallest denormal
// 	// {"5e-324", "5e-324"},
// 	// {"4e-324", "5e-324"},
// 	// {"3e-324", "5e-324"},
// 	// // too small
// 	// {"2e-324", "0"},
// 	// // way too small
// 	// {"1e-350", "0"},
// 	// {"1e-400000", "0"},

// 	// // try to overflow exponent
// 	// {"1e-4294967296", "0"},
// 	// {"1e-18446744073709551616", "0"},

// 	// // https://www.exploringbinary.com/java-hangs-when-converting-2-2250738585072012e-308/
// 	// {"2.2250738585072012e-308", "2.2250738585072014e-308"},
// 	// // https://www.exploringbinary.com/php-hangs-on-numeric-value-2-2250738585072011e-308/
// 	// {"2.2250738585072011e-308", "2.225073858507201e-308"},

// 	// // A very large number (initially wrongly parsed by the fast algorithm).
// 	// {"4.630813248087435e+307", "4.630813248087435e+307"},

// 	// // A different kind of very large number.
// 	// {"22.222222222222222", "22.22222222222222"},
// 	// {"2." + strings.Repeat("2", 4000) + "e+1", "22.22222222222222"},

// 	// // Exactly halfway between 1 and math.Nextafter(1, 2).
// 	// // Round to even (down).
// 	// {"1.00000000000000011102230246251565404236316680908203125", "1"},
// 	// // Slightly lower; still round down.
// 	// {"1.00000000000000011102230246251565404236316680908203124", "1"},
// 	// // Slightly higher; round up.
// 	// {"1.00000000000000011102230246251565404236316680908203126", "1.0000000000000002"},
// 	// // Slightly higher, but you have to read all the way to the end.
// 	// {"1.00000000000000011102230246251565404236316680908203125" + strings.Repeat("0", 10000) + "1", "1.0000000000000002"},

// 	// // Halfway between x := math.Nextafter(1, 2) and math.Nextafter(x, 2)
// 	// // Round to even (up).
// 	// {"1.00000000000000033306690738754696212708950042724609375", "1.0000000000000004"},
// }

// func TestGetFloatMore(t *testing.T) {
// 	for _, tc := range atofTests {
// 		t.Run(tc.in, func(t *testing.T) {
// 			f, _, _ := pauper.GetFloat([]byte(tc.in))
// 			if f != tc.out {
// 				t.Errorf("got %g instead of %g", f, tc.out)
// 			}
// 		})
// 	}
// }

// func BenchmarkGetFloat(b *testing.B) {
// 	buf := []byte(`-12345678900000000000000000000000000000000000000000000000000000000000000000000000000000000000.123E-10, "hi"`)
// 	for i := 0; i < b.N; i++ {
// 		pauper.GetFloat(buf)
// 	}
// }
