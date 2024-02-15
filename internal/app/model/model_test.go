package model

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"regexp"
	"strconv"
	"testing"
)

func Test_getHash(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: `Test getHash function - compare with reference value
send: "" - empty string
got: hash`,
		args: struct{ b []byte }{b: nil},
		want: "00000000",
	}, {
		name: `Test getHash function - compare with reference value
send: "https://yandex.ru/"
got: hash`,
		args: struct{ b []byte }{b: []byte("https://yandex.ru/")},
		want: "2eb63f75",
	}, {
		name: `Test getHash function - compare with reference value
send: "https://go.dev/"
got: hash`,
		args: struct{ b []byte }{b: []byte("https://go.dev/")},
		want: "a7930003",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHash(tt.args.b); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func BenchmarkGetHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getHash([]byte(strconv.Itoa(i)))
	}
}

func FuzzGetHash(f *testing.F) {
	m := make(map[string][]byte)
	f.Fuzz(func(t *testing.T, s string) {
		if regexp.MustCompile(`\w+`).FindString(s) != s {
			t.SkipNow()
		}
		h := getHash([]byte(s))
		if _, ok := m[h]; ok && !bytes.Equal(m[h], []byte(s)) {
			t.Error(s, string(m[h]), h)
		}
		m[h] = []byte(s)
	})
}
