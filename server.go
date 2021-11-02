package jsonrpc

import (
	"context"
	"encoding/json"
	"sync"
	"github.com/opentracing/opentracing-go"
)

type serverRequest struct {
	Version string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
	ID      *json.RawMessage `json:"id,omitempty"`
}

type serverResponse struct {
	Version string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

func (r *serverRequest) validate() bool {
	if r.Version != `2.0` {
		return false
	}
	if r.Method == `` {
		return false
	}
	return true
}

func (r *serverResponse) validate() bool {
	if r.Version != `2.0` {
		return false
	}
	return true
}

type Handler func(ctx context.Context, message *json.RawMessage) (interface{}, *Error)

type Server struct {
	methods map[string]Handler
}

func NewServer(methods map[string]Handler) *Server {
	server := &Server{
		methods: methods,
	}
	return server
}

func (s *Server) Call(ctx context.Context, data []byte) (resp []byte) {
	req, e := s.parseRequest(data)
	if e != nil {
		var batchData []json.RawMessage
		err := json.Unmarshal(data, &batchData)
		if err == nil {
			return s.batch(ctx, batchData)
		}

		return s.newErrorResponseBytes(e, nil)
	}

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetOperationName("handle rpc: "+req.Method)
		span.SetTag("rpc.method", req.Method)

		if req.ID != nil {
			if id, err := req.ID.MarshalJSON(); len(id) > 0 && err == nil {
				span.SetTag("rpc.request_id", string(id))
			}
		}
	}

	result, e := s.execute(ctx, req.Method, req.Params)
	if e != nil {
		return s.newErrorResponseBytes(e, req.ID)
	}

	return s.newResponseBytes(result, req.ID)
}

func (s *Server) batch(ctx context.Context, data []json.RawMessage) []byte {
	var wg sync.WaitGroup
	resp := make([]json.RawMessage, len(data))
	for i, d := range data {
		wg.Add(1)
		go func(i int, d json.RawMessage) {
			defer wg.Done()
			// TODO: add opentracing for batch requests
			resp[i] = s.Call(ctx, d)
		}(i, d)
	}
	wg.Wait()
	r, err := json.Marshal(resp)
	if err != nil {
		return s.newErrorResponseBytes(ErrorInternal, nil)
	}
	return r
}

func (s *Server) parseRequest(data []byte) (*serverRequest, *Error) {
	var req serverRequest
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, ErrorParse
	}
	if !req.validate() {
		return nil, ErrorRequest
	}

	return &req, nil
}

func (s *Server) execute(ctx context.Context, method string, params *json.RawMessage) (result interface{}, e *Error) {
	defer func() {
		if r := recover(); r != nil {
			e = ErrorInternal
		}
	}()

	if _, ok := s.methods[method]; !ok {
		return nil, ErrorMethod
	}

	f := s.methods[method]
	result, e = f(ctx, params)

	return result, e
}

func (s *Server) newResponseBytes(result interface{}, id *json.RawMessage) []byte {
	resp := &serverResponse{
		Version: `2.0`,
		ID:      id,
		Result:  result,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return s.newErrorResponseBytes(ErrorInternal, id)
	}
	return data
}

func (s *Server) newErrorResponseBytes(e *Error, id *json.RawMessage) []byte {
	resp := &serverResponse{
		Version: `2.0`,
		ID:      id,
		Error:   e,
	}
	data, _ := json.Marshal(resp)
	return data
}
