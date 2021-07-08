package sinister

import (
	"encoding/json"
	"net"
	"net/http"

	"go.uber.org/zap"
)

type HC struct {
	w      http.ResponseWriter
	r      *http.Request
	logger *zap.Logger
	params []*Param
}

func (hc *HC) set(w http.ResponseWriter, r *http.Request, logger *zap.Logger, params []*Param) {
	hc.w = w
	hc.r = r
	hc.logger = logger
	hc.params = params
}

func (hc *HC) reset() {
	hc.w = nil
	hc.r = nil
	hc.logger = nil
	hc.params = nil
}

func newHC() *HC {
	return &HC{
		w:      nil,
		r:      nil,
		logger: nil,
		params: nil,
	}
}

// JSONS ...
func (hc *HC) JSONS(code int, data string) {
	// r := &httpResponse{code, msg}
	r := newHTTPResponse(code, data)
	hc.w.WriteHeader(code)
	if err := json.NewEncoder(hc.w).Encode(r); err != nil {
		http.Error(hc.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (hc *HC) JSONI(code int, data interface{}) {
	hc.w.WriteHeader(code)
	if err := json.NewEncoder(hc.w).Encode(data); err != nil {
		http.Error(hc.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (hc *HC) RAWS(code int, data string) {
	_, err := hc.w.Write([]byte(data))
	if err != nil {
		http.Error(hc.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (hc *HC) RAWB(code int, data []byte) {
	_, err := hc.w.Write(data)
	if err != nil {
		http.Error(hc.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (hc *HC) Param(param string) URLParam {
	if len(hc.params) == 0 {
		return ""
	}
	return URLParam(findParam(hc.params, param))
}

func (hc *HC) Log(msg string, level LogLevel) {
	switch level {
	case DEBUG:
		hc.logger.Debug(msg)
	case ERROR:
		hc.logger.Error(msg)
	case INFO:
		hc.logger.Info(msg)
	case FATAL:
		hc.logger.Fatal(msg)
	case WARN:
		hc.logger.Warn(msg)
	}
}

func (hc *HC) Query(key string) (string, error) {
	if hc.r.URL.Query().Get(key) == "" {
		return "", ErrQueryNotFound
	}
	return hc.r.URL.Query().Get(key), nil
}
func (hc *HC) Cookie(name string) (*http.Cookie, error) {
	c, err := hc.r.Cookie(name)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (hc *HC) MIME(mime MIME) {
	hc.w.Header().Set("Content-Type", string(mime))
}

func (hc *HC) ClientIP() string {
	ip, _, err := net.SplitHostPort(hc.r.RemoteAddr)
	if err != nil {
		forward := hc.r.Header.Get("X-Forwarded-For")
		return forward
	}
	return ip
}
