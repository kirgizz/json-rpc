package jsonrpc

import (
	"fmt"
)

const (
	ErrorCodeRequest  = -32600
	ErrorCodeMethod   = -32601
	ErrorCodeParams   = -32602
	ErrorCodeInternal = -32603
	ErrorCodeParse    = -32700
	ErrorCodeHandler  = -32000 // deprecated. Use ErrorCodeServer instead
	ErrorCodeServer   = -32000
)

var (
	ErrorParse    = NewError(ErrorCodeParse, "Parse error")
	ErrorRequest  = NewError(ErrorCodeRequest, "Invalid Request")
	ErrorMethod   = NewError(ErrorCodeMethod, "Method not found")
	ErrorParams   = NewError(ErrorCodeParams, "Invalid params")
	ErrorInternal = NewError(ErrorCodeInternal, "Internal JSON-RPC error")
)

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func NewErrorFromString(message string) *Error {
	return &Error{Code: ErrorCodeServer, Message: message}
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error [%d]: %s", e.Code, e.Message)
}
