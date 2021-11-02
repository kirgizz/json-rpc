package jsonrpc

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"gitlab.octafx.com/go-libs/tracing"
)

type HTTPServer struct {
	server *Server
	tracer opentracing.Tracer
	logger *zap.Logger
}

func NewHTTPServer(methods map[string]Handler, tracer opentracing.Tracer, logger *zap.Logger) *HTTPServer {
	if tracer == nil {
		tracer = opentracing.NoopTracer{}
	}

	httpServer := &HTTPServer{
		server: NewServer(methods),
		tracer: tracer,
		logger: logger,
	}

	return httpServer
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	spanCtx, err := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		s.logger.Warn("can't extract span context from headers", zap.Error(err))
	}

	span := s.tracer.StartSpan("handle rpc request", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	span.SetTag("rpc.endpoint", r.Host + r.URL.Path)
	ext.HTTPMethod.Set(span, r.Method)
	tracing.ProcessTraceId(span)

	ctx := opentracing.ContextWithSpan(r.Context(), span)

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
