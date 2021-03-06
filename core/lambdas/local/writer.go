package local

import (
	"net/http"
)

const cacheSize = 1024

func wrapWriter(w http.ResponseWriter, headers map[string]string) *cachedWriter {
	return &cachedWriter{wrap: w, headers: headers}
}

type cachedWriter struct {
	offset  int
	buffer  [cacheSize]byte
	flushed bool
	headers map[string]string
	wrap    http.ResponseWriter
}

func (cw *cachedWriter) Write(p []byte) (int, error) {
	if cw.flushed {
		return cw.wrap.Write(p)
	}
	if cw.offset+len(p) > len(cw.buffer) {
		_, err := cw.flush()
		if err != nil {
			return 0, err
		}
		return cw.wrap.Write(p)
	}
	copy(cw.buffer[cw.offset:], p)
	cw.offset += len(p)
	return len(p), nil
}

func (cw *cachedWriter) flush() (int, error) {
	for k, v := range cw.headers {
		cw.wrap.Header().Set(k, v)
	}

	n, err := cw.wrap.Write(cw.buffer[:cw.offset])
	if err != nil {
		return 0, err
	}
	cw.flushed = true
	return n, nil
}
