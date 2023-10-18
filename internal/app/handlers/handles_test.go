package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"shortener/config"
	"shortener/internal/app/model"
	strg "shortener/internal/app/storage"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
)

var _ = func() bool {
	testing.Init()
	return true
}()

var storage = strg.New()
var api = NewAPI(storage)

func Test_getHash(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: `"Hash: "" - empty string"`,
		args: struct{ b []byte }{b: nil},
		want: "389589f3",
	}, {
		name: `Hash: "https://yandex.ru/"`,
		args: struct{ b []byte }{b: []byte("https://yandex.ru/")},
		want: "1e320d4f",
	}, {
		name: `Hash: "https://go.dev/"`,
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

func TestHandlerAndStorage(t *testing.T) {
	mem := make(map[string]string)
	r := chi.NewRouter()
	r.Post("/", api.PostPlainText)
	r.Get("/{sn}", api.GetRoot)
	r.NotFound(api.Default)
	r.MethodNotAllowed(api.Default)
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
		name:       `Success: POST "https://go.dev/" -> hash("https://go.dev/")`,
		args:       args{method: http.MethodPost, path: "/", body: strings.NewReader("https://go.dev/")},
		memSlot:    "1",
		wantStatus: http.StatusCreated,
		wantBody:   []byte("e4546b92"),
	}, {
		name:       `Success: GET hash("https://go.dev/") -> "https://go.dev/"`,
		args:       args{method: http.MethodGet, path: "1", body: nil},
		wantStatus: http.StatusTemporaryRedirect,
		wantBody:   []byte("https://go.dev/"),
	}, {
		name:       "Wrong PUT: rejected by router",
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

func TestPostPlainText(t *testing.T) {
	s := strg.New()
	a := NewAPI(s)
	type args struct {
		body io.Reader
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{{
		name:       `Success POST "https://go.dev/"`,
		args:       args{body: strings.NewReader("https://go.dev/")},
		wantStatus: http.StatusCreated,
	}, {
		name:       `#1 Error on io.ReadALL `,
		args:       args{body: strings.NewReader("https://go.dev/")},
		wantStatus: http.StatusBadRequest,
	}, {
		name:       `#2 Error on url.Join `,
		args:       args{body: strings.NewReader("https://go.dev/")},
		wantStatus: http.StatusBadRequest,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch {
			case strings.HasPrefix(tt.name, "#1"):
				tt.args.body = iotest.ErrReader(errors.New("io error occurred"))
			case strings.HasPrefix(tt.name, "#2"):
				wk := config.BaseURL
				defer func() {
					config.BaseURL = wk
				}()
				str := "ht\tp://wewe.we/"
				config.BaseURL = &str
			}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://localhost/", tt.args.body)
			a.PostPlainText(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if resp.StatusCode == http.StatusCreated {
				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				_, ok := storage.Load(string(b))
				assert.True(t, ok)
			}
		})
	}
}

func TestPostJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       io.Reader
		wantStatus int
		wantBody   string
	}{{
		name:       "Success",
		body:       strings.NewReader(`{"url":"http://www.yandex.ru"}`),
		wantStatus: http.StatusCreated,
		wantBody:   `{"result":"http://localhost:8080/86e99165"}`,
	}, {
		name:       "Bad JSON",
		body:       strings.NewReader(`{url":"http://www.yandex.ru"}`),
		wantStatus: http.StatusBadRequest,
		wantBody:   ``,
	}, {
		name:       "#1 io error on json.decode",
		body:       strings.NewReader(`{url":"http://www.yandex.ru"}`),
		wantStatus: http.StatusBadRequest,
		wantBody:   ``,
	}, {
		name:       "#2 Error on url.Join",
		body:       strings.NewReader(`{url":"http://www.yandex.ru"}`),
		wantStatus: http.StatusBadRequest,
		wantBody:   ``,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch {
			case strings.HasPrefix(tt.name, "#1"):
				tt.body = iotest.ErrReader(errors.New("io error occurred"))
			case strings.HasPrefix(tt.name, "#2"):
				wk := config.BaseURL
				defer func() {
					config.BaseURL = wk
				}()
				str := "ht\tp://wewe.we/"
				config.BaseURL = &str
			}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://localhost/", tt.body)
			api.PostJSON(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if resp.StatusCode == http.StatusCreated {
				var mr model.Response
				err := json.NewDecoder(resp.Body).Decode(&mr)
				require.NoError(t, err)
				_, ok := storage.Load(mr.Result)
				assert.True(t, ok)
			}
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	s := strg.New()
	a := NewAPI(s)
	srv := httptest.NewServer(http.HandlerFunc(a.GetUserURLs))
	defer srv.Close()
	tests := []struct {
		name       string
		wantStatus int
		wantBody   []byte
	}{{
		name:       "Empty storage",
		wantStatus: http.StatusNoContent,
		wantBody:   []byte(""),
	}, {
		name:       "One_record_not_work",
		wantStatus: http.StatusNoContent,
		wantBody:   []byte(""),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(srv.URL)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, b)
			u := uuid.NewString()
			s.Store(u, "ok")
		})
	}
}

//func TestGetByShortName(t *testing.T) {
//	r := chi.NewRouter()
//	r.Get("/{sn}", api.GetRoot)
//	srv := httptest.NewServer(r)
//	defer srv.Close()
//	//storage.Add(*config.BaseURL+"1", "https://go.dev/")
//	type args struct {
//		path string
//		body io.Reader
//	}
//	tests := []struct {
//		name       string
//		args       args
//		memSlot    string
//		wantStatus int
//		wantBody   []byte
//	}{{
//		name:       "Success GET",
//		args:       args{path: "/1", body: nil},
//		wantStatus: http.StatusTemporaryRedirect,
//		wantBody:   []byte("https://go.dev/"),
//	}, {
//		name:       "Wrong GET",
//		args:       args{path: "/0", body: nil},
//		wantStatus: http.StatusBadRequest,
//		wantBody:   []byte("not found\n"),
//	}}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			req, err := http.NewRequest(http.MethodGet, srv.URL+tt.args.path, tt.args.body)
//			require.NoError(t, err)
//			cl := http.Client{
//				Transport: nil,
//				CheckRedirect: func(req *http.Request, via []*http.Request) error {
//					return http.ErrUseLastResponse
//				},
//				Jar:     nil,
//				Timeout: 0,
//			}
//			resp, err := cl.Do(req)
//			require.NoError(t, err)
//			defer resp.Body.Close()
//			assert.Equal(t, tt.wantStatus, resp.StatusCode)
//			b, err := io.ReadAll(resp.Body)
//			require.NoError(t, err)
//			assert.Equal(t, tt.wantBody, b)
//		})
//	}
//}

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
