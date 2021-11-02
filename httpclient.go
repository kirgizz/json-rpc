package jsonrpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

type HTTPClient struct {
	url     string
	headers map[string]string
}

func NewHTTPClient(url string) *HTTPClient {
	client := &HTTPClient{
		url: url,
	}
	return client
}

func (c *HTTPClient) SetHeaders(headers map[string]string) {
	c.headers = headers
}

func (c *HTTPClient) Call(method string, params interface{}) (*Response, error) {
	id, err := c.generateID()
	if err != nil {
		return nil, err
	}
	body, err := NewRequestBytes(method, params, id)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(`Content-Type`, `application/json`)
	req.Header.Set(`Accept`, `application/json`)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(`invalid status code: %v`, resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response, err := ParseResponse(data)
	if err != nil {
		return nil, err
	}

	if !response.validate() {
		return nil, fmt.Errorf(`invalid response: %s`, data)
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return response, nil
}

func (c *HTTPClient) CallBatch(requests []*Request) (Responses, error) {
	var err error
	for _, req := range requests {
		req.ID, err = c.generateID()
		if err != nil {
			return nil, err
		}
	}

	body, err := NewBatchRequestBytes(requests)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(`Content-Type`, `application/json`)
	req.Header.Set(`Accept`, `application/json`)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(`invalid status code: %v`, resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responses, err := ParseBatchResponse(requests, data)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

func (c *HTTPClient) generateID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return `0`, err
	}
	id := fmt.Sprintf(`%x-%x-%x-%x-%x`, b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return id, nil
}
