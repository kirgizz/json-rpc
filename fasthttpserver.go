package jsonrpc

import (
	"context"
	"github.com/valyala/fasthttp"
)

type FastHTTPServer struct {
	server *Server
}

func NewFastHTTPServer(methods map[string]Handler) *FastHTTPServer {
	httpServer := &FastHTTPServer{
		server: NewServer(methods),
	}
	return httpServer
}

func (s *FastHTTPServer) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()

	var resp []byte
	if !ctx.IsPost() || len(body) == 0 {
		resp = s.server.newErrorResponseBytes(ErrorParse, nil)
	} else {
		resp = s.server.Call(context.Background(), body)
	}
	ctx.SetConnectionClose()
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(resp)
}
