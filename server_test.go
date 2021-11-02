package jsonrpc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"context"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
)

func TestCall_CorrectEmptyResponse(t *testing.T) {
	s := NewServer(map[string]Handler{
		"foo": func(ctx context.Context, message *json.RawMessage) (interface{}, *Error) {
			return nil, nil
		},
	})

	resp := s.Call(
		context.TODO(),
		[]byte(`{ "method": "foo", "id": "123", "jsonrpc": "2.0"}`),
	)

	assert.Equal(t, `{"jsonrpc":"2.0","id":"123"}`, string(resp))
}

func TestCall_CorrectStringResponse(t *testing.T) {
	s := NewServer(map[string]Handler{
		"foo": func(ctx context.Context, message *json.RawMessage) (interface{}, *Error) {
			return "correct response", nil
		},
	})

	resp := s.Call(
		context.TODO(),
		[]byte(`{ "method": "foo", "id": "123", "jsonrpc": "2.0"}`),
	)

	assert.Equal(t, `{"jsonrpc":"2.0","id":"123","result":"correct response"}`, string(resp))
}

func TestCall_CorrectStructResponse(t *testing.T) {
	s := NewServer(map[string]Handler{
		"foo": func(ctx context.Context, message *json.RawMessage) (interface{}, *Error) {
			return struct {
				Foo string `json:"foo"`
			}{
				Foo : "bar",
			}, nil
		},
	})

	resp := s.Call(
		context.TODO(),
		[]byte(`{ "method": "foo", "id": "123", "jsonrpc": "2.0"}`),
	)

	assert.Equal(t, `{"jsonrpc":"2.0","id":"123","result":{"foo":"bar"}}`, string(resp))
}

func TestCall_HandleIncorrectMethodName(t *testing.T) {
	s := NewServer(map[string]Handler{
		"foo": func(ctx context.Context, message *json.RawMessage) (interface{}, *Error) {
			return "response", nil
		},
	})

	resp := s.Call(
		context.TODO(),
		[]byte(`{ "method": "incorrect_method", "id": "123", "jsonrpc": "2.0"}`),
	)

	assert.Equal(t, `{"jsonrpc":"2.0","id":"123","error":{"code":-32601,"message":"Method not found"}}`, string(resp))
}

func TestCall_CorrectWorkWithSpan(t *testing.T) {
	s := NewServer(map[string]Handler{
		"foo": func(ctx context.Context, message *json.RawMessage) (interface{}, *Error) {
			return "response", nil
		},
	})

	span := (&opentracing.NoopTracer{}).StartSpan("foo1")
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	resp := s.Call(
		ctx,
		[]byte(`{ "method": "foo", "id": "123", "jsonrpc": "2.0"}`),
	)

	assert.Equal(t, `{"jsonrpc":"2.0","id":"123","result":"response"}`, string(resp))

	// TODO: check that span has tags: rpc.method and rpc.ID
}
