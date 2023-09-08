package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"path"
	"sync"
)

var storage = sync.OnceValue(func() map[string][]byte {
	return make(map[string][]byte)
})

func getHash(b []byte) string {
	h := md5.Sum(b)
	d := make([]byte, len(h)/4)
	for i := range d {
		d[i] = h[i] + h[i+len(h)/4] + h[i+len(h)/2] + h[i+3*len(h)/4]
	}
	return hex.EncodeToString(d)
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost &&
		r.URL.Path == "/" {
		b, e := io.ReadAll(r.Body)
		if e != nil {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
		h := getHash(b)
		storage()[h] = b
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://" + r.Host + "/" + h))
		return
	} else if d, f := path.Split(r.URL.Path); r.Method == http.MethodGet &&
		d == "/" && f != "" {
		b, ok := storage()[f]
		if !ok {
			http.Error(w, f, http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", string(b))
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write(b)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
