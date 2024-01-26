package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/Stas9132/shortener/config"
	"github.com/Stas9132/shortener/internal/app/handlers/middleware"
	"github.com/Stas9132/shortener/internal/app/model"
	strg "github.com/Stas9132/shortener/internal/app/storage"
	"github.com/Stas9132/shortener/internal/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = func() bool {
	testing.Init()
	return true
}()

var storage, _ = strg.NewFileStorage(context.Background(), logger.NewDummy())
var api = NewAPI(context.Background(), logger.NewDummy(), storage)

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
		name: ` Create new short URL
send correct URL
got: 
- response shorturl
- status created`,
		args:       args{method: http.MethodPost, path: "/", body: strings.NewReader("https://go.dev/")},
		memSlot:    "1",
		wantStatus: http.StatusCreated,
		wantBody:   []byte("a7930003"),
	}, {
		name: `Get original URL from short
send: 
- correct URL
- shortURL from previous step
got: 
- response original URL
- status TemporaryRedirect`,
		args:       args{method: http.MethodGet, path: "1", body: nil},
		wantStatus: http.StatusTemporaryRedirect,
		wantBody:   []byte("https://go.dev/"),
	}, {
		name: `Wrong PUT: rejected by router
send request with wrong method
got status BadRequest`,
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
	s, _ := strg.NewFileStorage(context.Background(), logger.NewDummy())
	a := NewAPI(context.Background(), logger.NewDummy(), s)
	type args struct {
		body io.Reader
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{{
		name: `Success POST: 
send correct URL
got status Created`,
		args:       args{body: strings.NewReader("https://go.dev/")},
		wantStatus: http.StatusCreated,
	}, {
		name: `#1 Error on io.ReadALL
send bad request
set io.error
got status BadRequest`,
		args:       args{body: strings.NewReader("https://go.dev/")},
		wantStatus: http.StatusBadRequest,
	}, {
		name: `#2 Error on url.Join
send correct URL
set bad config variables
got status BadRequest`,
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
			middleware.Authorization(http.HandlerFunc(a.PostPlainText)).ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if resp.StatusCode == http.StatusCreated {
				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				_, ok := s.Load(string(b))
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
		name: `Success
send correct request
got
- status Created
- shortURL`,
		body:       strings.NewReader(`{"url":"http://www.yandex.ru"}`),
		wantStatus: http.StatusCreated,
		wantBody:   `{"result":"http://localhost:8080/86e99165"}`,
	}, {
		name: `Bad JSON
send correct request
got status BadRequest`,
		body:       strings.NewReader(`{url":"http://www.yandex.ru"}`),
		wantStatus: http.StatusBadRequest,
		wantBody:   ``,
	}, {
		name: `#1 io error on json.decode
send correct url
set io.error
got status BadRequest`,
		body:       strings.NewReader(`{url":"http://www.yandex.ru"}`),
		wantStatus: http.StatusBadRequest,
		wantBody:   ``,
	}, {
		name: `#2 Error on url.Join
send correct url
set bad config variables
got status BadRequest`,
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
	s, _ := strg.NewFileStorage(context.Background(), logger.NewDummy())
	a := NewAPI(context.Background(), logger.NewDummy(), s)
	srv := httptest.NewServer(http.HandlerFunc(a.GetUserURLs))
	defer srv.Close()
	tests := []struct {
		name       string
		wantStatus int
	}{{
		name: `Empty storage
send get request
got status NoContent`,
		wantStatus: http.StatusNoContent,
	}, {
		name: `One_record
send get request
got 
status OK
list with one record`,
		wantStatus: http.StatusOK,
	}, {
		name: `Two record
send get request
got 
status OK
list with two record`,
		wantStatus: http.StatusOK,
	}}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, err)
			req.Header.Set("Accept-Encoding", "identity")
			resp, err := (&http.Client{}).Do(req)
			s.Store(uuid.NewString(), "ok")
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			if resp.StatusCode == http.StatusOK {
				var v []any
				err = json.NewDecoder(resp.Body).Decode(&v)
				require.NoError(t, err)
				require.Equal(t, i, len(v))
			}
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
