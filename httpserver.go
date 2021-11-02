package jsonrpc

import (
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type HTTPServer struct {
	server *Server
	logger *zap.Logger
}

func NewHTTPServer(methods map[string]Handler, logger *zap.Logger) *HTTPServer {

	httpServer := &HTTPServer{
		server: NewServer(methods),
		logger: logger,
	}

	return httpServer
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()


	ctx := opentracing.ContextWithSpan(r.Context(), nil)

	body, err := ioutil.ReadAll(r.Body)
	var resp []byte
	if err != nil || len(body) == 0 {
		resp = s.server.newErrorResponseBytes(ErrorParse, nil)
	} else {
		resp = s.server.Call(ctx, body)
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}
