# jsonrpc

JSON RPC 2.0 Client and Server

## Server example

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.octafx.com/go-libs/jsonrpc"
	"net/http"
)

type SumParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type SumResult struct {
	Sum int `json:"sum"`
}

func Sum(_ context.Context, message *json.RawMessage) (interface{}, *jsonrpc.Error) {
	var params SumParams
	err := json.Unmarshal(*message, &params)
	if err != nil {
		return nil, jsonrpc.ErrorParams
	}

	result := &SumResult{
		Sum: params.A + params.B,
	}

	return result, nil
}

func main() {
	methods := map[string]jsonrpc.Handler {
		`test.sum`: Sum,
	}
	server := jsonrpc.NewHTTPServer(methods, tracer)

	http.Handle(`/`, server)

	err := http.ListenAndServe(`:4000`, nil)
	if err != nil {
		fmt.Printf(`http listen error: %v`, err)
	}
}
```

## Client example

```go
package main

import (
	"fmt"
	"gitlab.octafx.com/go-libs/jsonrpc"
)

type SumParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type SumResult struct {
	Sum int `json:"sum"`
}

func main() {
	client := jsonrpc.NewHTTPClient(`http://127.0.0.1:4000/`)

	response, err := client.Call(`test.sum`, SumParams{A: 11, B: 22})
	if err != nil {
		switch e := err.(type) {
		case *jsonrpc.Error:
			fmt.Printf("server returns jsonrpc error: code=%v, message=%v\n", e.Code, e.Message)
		default:
			fmt.Printf("client error: %v\n", err)
		}
		return
	}

	var result SumResult
	if err = response.GetResult(&result); err != nil {
		fmt.Printf("result decode error: %v\n", err)
		return
	}
	fmt.Printf("result is %v\n", result.Sum)
}
```


## Client batch request example

```go
package main

import (
	"fmt"
	"gitlab.octafx.com/go-libs/jsonrpc"
)

type SumParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type SumResult struct {
	Sum int `json:"sum"`
}

func main() {
	client := jsonrpc.NewHTTPClient(`http://127.0.0.1:4000/`)

	var result SumResult

	requests := []*jsonrpc.Request{
		jsonrpc.NewRequest(`test.sum`, SumParams{A: 2, B: 3}),
		jsonrpc.NewRequest(`test.sum`, SumParams{A: 5, B: 7}),
		jsonrpc.NewRequest(`test.sum`, SumParams{A: 27, B: 15}),
	}
	responses, err := client.CallBatch(requests)
	if err != nil {
		fmt.Printf("batch request error: %v\n", err)
		return
	}

	// find result by index
	response1 := responses[0]
	if response1.Error != nil {
		fmt.Printf("batch error for request 1: %v\n", err)
		return
	}
	if err = response1.GetResult(&result); err != nil {
		fmt.Printf("result decode error: %v", err)
		return
	}
	fmt.Printf("batch result 1 is %v\n", result.Sum)

	// find result by request
	response2 := responses.ByRequest(requests[1])
	if response2.Error != nil {
		fmt.Printf("batch error for request 2: %v\n", err)
		return
	}
	if err = response2.GetResult(&result); err != nil {
		fmt.Printf("result decode error: %v\n", err)
		return
	}
	fmt.Printf("batch result 2 is %v\n", result.Sum)
}
```

## FastHTTP Server example

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"gitlab.octafx.com/go-libs/jsonrpc"
)

type SumParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type SumResult struct {
	Sum int `json:"sum"`
}

func Sum(_ context.Context, message *json.RawMessage) (interface{}, *jsonrpc.Error) {
	var params SumParams
	err := json.Unmarshal(*message, &params)
	if err != nil {
		return nil, jsonrpc.ErrorParams
	}

	result := &SumResult{
		Sum: params.A + params.B,
	}

	return result, nil
}

func main() {
	methods := map[string]jsonrpc.Handler{
		`test.sum`: Sum,
	}
	server := jsonrpc.NewFastHTTPServer(methods)

	err := fasthttp.ListenAndServe(":4000", server.HandleFastHTTP)
	if err != nil {
		fmt.Printf(`http listen error: %v`, err)
	}
}
```

## Middleware example

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.octafx.com/go-libs/jsonrpc"
	"net"
	"net/http"
)

type SumParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type SumResult struct {
	Sum int `json:"sum"`
	IP string `json:"ip"`
}

func Sum(ctx context.Context, message *json.RawMessage) (interface{}, *jsonrpc.Error) {
	var params SumParams
	err := json.Unmarshal(*message, &params)
	if err != nil {
		return nil, jsonrpc.ErrorParams
	}

	result := &SumResult{
		Sum: params.A + params.B,
		IP: ctx.Value("ip").(string),
	}

	return result, nil
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ip == `1.2.3.4` {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "ip", ip)))
	})
}

func main() {
	methods := map[string]jsonrpc.Handler {
		`test.sum`: Sum,
	}
	server := jsonrpc.NewHTTPServer(methods)

	http.Handle(`/`, middleware(server))

	err := http.ListenAndServe(`:4000`, nil)
	if err != nil {
		fmt.Printf(`http listen error: %v`, err)
	}
}
```
