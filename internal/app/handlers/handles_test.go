package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
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
		name: "Empty string",
		args: struct{ b []byte }{b: nil},
		want: "389589f3",
	}, {
		name: "yandex.ru",
		args: struct{ b []byte }{b: []byte("https://yandex.ru/")},
		want: "1e320d4f",
	}, {
		name: "go.dev",
		args: struct{ b []byte }{b: []byte("https://go.dev/")},
		want: "e4546b92",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHash(tt.args.b); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMainHandler(t *testing.T) {
	mem := make(map[string]string)
	r := chi.NewRouter()
	r.Post("/", MainPage)
	r.Get("/{sn}", GetByShortName)
	r.NotFound(Default)
	r.MethodNotAllowed(Default)
	srv := httptest.NewServer(r)
	defer srv.Close()
	type args struct {
		method string
		path   string
		body   io.Reader
	}
	tests := []struct {
		name       string
		args       args
		memSlot    string
		wantStatus int
		wantBody   []byte
	}{{
		name:       "Success POST",
		args:       args{method: http.MethodPost, path: "/", body: strings.NewReader("https://go.dev/")},
		memSlot:    "1",
		wantStatus: http.StatusCreated,
		wantBody:   []byte("e4546b92"),
	}, {
		name:       "Success GET",
		args:       args{method: http.MethodGet, path: "1", body: nil},
		wantStatus: http.StatusTemporaryRedirect,
		wantBody:   []byte("https://go.dev/"),
	}, {
		name:       "Wrong PUT",
		args:       args{method: http.MethodPut, path: "/", body: nil},
		wantStatus: http.StatusBadRequest,
		wantBody:   []byte{},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := strconv.Atoi(tt.args.path); err == nil {
				tt.args.path = mem[tt.args.path]
			}
			req, err := http.NewRequest(tt.args.method, srv.URL+tt.args.path, tt.args.body)
			require.NoError(t, err)
			cl := http.Client{
				Transport: nil,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
				Jar:     nil,
				Timeout: 0,
			}
			resp, err := cl.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			m := regexp.MustCompile(`.*//.*/(\w{8})`).FindSubmatch(b)
			if len(m) == 2 {
				assert.Equal(t, tt.wantBody, m[1])
				mem[tt.memSlot] = "/" + string(m[1])
			} else {
				assert.Equal(t, tt.wantBody, b)
			}
		})
	}
}
