package gzip

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockResponseWriter struct {
	callCheck       map[string]struct{}
	retHeader       http.Header
	passWrite       []byte
	retWriteI       int
	retWriteE       error
	passWriteHeader int
}

func (w *mockResponseWriter) Header() http.Header {
	w.callCheck["Header"] = struct{}{}
	return w.retHeader
}

func (w *mockResponseWriter) Write(b []byte) (int, error) {
	w.callCheck["Write"] = struct{}{}
	w.passWrite = b
	return w.retWriteI, w.retWriteE
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.passWriteHeader = statusCode
	w.callCheck["WriteHeader"] = struct{}{}
}

func TestCompressWriter(t *testing.T) {
	tests := []struct {
		name  string
		call  string
		call2 string
		tw    *mockResponseWriter
		b     []byte
		sc    int
	}{{name: `nocall`,
		call: "no", tw: &mockResponseWriter{
			callCheck: make(map[string]struct{}),
		}}, {
		name: `Call Header
check header called on mocked object`,
		call: "Header",
		tw: &mockResponseWriter{
			callCheck: make(map[string]struct{}),
			retHeader: http.Header{},
		}}, {
		name: `Call Write
check Write called on mocked object`,
		call: "Write",
		tw: &mockResponseWriter{
			callCheck: make(map[string]struct{}),
			passWrite: make([]byte, 100),
		}}, {
		name: `Call WriteHeader httpStatusOk 
check WriteHeader & Header called on mocked object`,
		call:  "WriteHeader",
		call2: "Header",
		sc:    http.StatusOK,
		tw: &mockResponseWriter{
			callCheck: make(map[string]struct{}),
			retHeader: http.Header{},
		}}, {
		name: `Call WriteHeader httpStatusBadRequest 
check WriteHeader called on mocked object`,
		call: "WriteHeader",
		sc:   http.StatusBadRequest,
		tw: &mockResponseWriter{
			callCheck: make(map[string]struct{}),
			retHeader: http.Header{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw := newCompressWriter(tt.tw)
			var ok2 bool
			switch {
			case tt.call == "Header":
				cw.Header()
			case tt.call == "Write":
				_, _ = cw.Write(tt.b)
				_ = cw.Close()
			case tt.call == "WriteHeader":
				cw.WriteHeader(tt.sc)
				require.Equal(t, tt.sc, tt.tw.passWriteHeader)
			default:
				ok2 = true
			}
			_, ok := tt.tw.callCheck[tt.call]
			require.True(t, ok || ok2)
			if len(tt.call2) > 0 {
				_, ok = tt.tw.callCheck[tt.call2]
				require.True(t, ok)
			}
		})
	}
}
