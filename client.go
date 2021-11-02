package jsonrpc

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      interface{} `json:"id,omitempty"`
}

type Response struct {
	Version string           `json:"jsonrpc"`
	ID      interface{}      `json:"id"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

type Responses []*Response

func NewRequest(method string, params interface{}) *Request {
	r := &Request{
		Method:  method,
		Params:  params,
		Version: `2.0`,
	}
	return r
}

func (r *Request) validate() bool {
	if r.Version != `2.0` {
		return false
	}
	if r.Method == `` {
		return false
	}
	return true
}

func (r *Response) validate() bool {
	return r.Version == `2.0`
}

func (r *Response) GetResult(v interface{}) error {
	return json.Unmarshal(*r.Result, v)
}

func (responses Responses) ByRequest(request *Request) *Response {
	for _, resp := range responses {
		if resp.ID == request.ID {
			return resp
		}
	}
	return nil
}

func NewRequestBytes(method string, params interface{}, id interface{}) ([]byte, error) {
	req := &Request{
		Version: `2.0`,
		Method:  method,
		Params:  params,
		ID:      id,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return reqJson, nil
}

func NewBatchRequestBytes(requests []*Request) ([]byte, error) {
	reqJson, err := json.Marshal(requests)
	if err != nil {
		return nil, err
	}

	return reqJson, nil
}

func ParseResponse(data []byte) (*Response, error) {
	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	if !response.validate() {
		return nil, fmt.Errorf(`invalid response in batch`)
	}

	return &response, nil
}

func ParseBatchResponse(requests []*Request, data []byte) (Responses, error) {
	var responseSlice []Response
	err := json.Unmarshal(data, &responseSlice)
	if err != nil {
		return nil, err
	}

	responses := make([]*Response, len(requests))
	for i, _ := range responseSlice {
		for index, req := range requests {
			if req.ID == responseSlice[i].ID {
				responses[index] = &responseSlice[i]
			}
		}
	}

	for _, resp := range responses {
		if resp == nil {
			return nil, fmt.Errorf(`response for batch request not returned`)
		}
		if !resp.validate() {
			return nil, fmt.Errorf(`invalid response in batch`)
		}
	}

	return responses, nil
}
