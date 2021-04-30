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
		`"hello"`:  "hello",
		`   "hi"`:  "hi",
		` "hi\""`:  "hi\"",
		`     ""`:  "",
		`"árbol"`:  "árbol",
		`"\u0020"`: " ",
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
	buf := []byte(`"1234567890qwertyuiopasdfghjklzxcvbnm áéíóú"`)
	for i := 0; i < b.N; i++ {
		pauper.GetString(buf)
	}
}
func BenchmarkGetString(b *testing.B) {
	buf := []byte(`"1234567890qwertyuiopasdfghjklzxcvbnm áéíóú \uD83D\uDCA9 \"💩\""`)
	//buf := []byte(`"1234567890qwertyuiopasdfghjklzxcvbnm áéíóú"`)
	for i := 0; i < b.N; i++ {
		pauper.GetString(buf)
	}
}

func TestGetNumber(t *testing.T) {
	for from, expected := range map[string]float64{
		"0,": 0,
		" -0,": 0,
		"10,": 10,
		"12345678900000000000000000000000000000000000000000000000000000000000000000000000000000000000,": 1.23456789e91,
		"1.23,": 1.23,
		"1.23e4,": 12300,
		"1.23e+4,": 12300,
		"123e4,": 1230000,
		"123e-2,": 1.23,
	} {
		t.Run(from, func(t *testing.T) {
			buf := []byte(from)
			got, i, err := pauper.GetNumber(buf)
			if err != nil {
				t.Error(err)
			}
			if buf[i] != ',' {
				t.Errorf("%q at %d?", buf, i)
			}
			if got != expected {
				t.Errorf("%g != %g", got, expected)
			}
		})
	}
}

func BenchmarkGetNumber(b *testing.B) {
	buf := []byte(`-12345678900000000000000000000000000000000000000000000000000000000000000000000000000000000000.123E-10, "hi"`)
	for i := 0; i < b.N; i++ {
		pauper.GetNumber(buf)
	}
}
