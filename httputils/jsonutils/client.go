package jsonutils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tukdesk/gopool"
)

const (
	codeNon = iota
	CodeConnectionError
	CodeReadResponseBodyError
	CodeJsonDecodeError
)

func newHTTPClient() (interface{}, error) {
	return &http.Client{}, nil
}

func newSimpleClientError(code int, err error) error {
	return NewAPIError(0, code, err.Error())
}

type SimpleClient struct {
	pool *gopool.Pool
}

func NewSimpleClient() (*SimpleClient, error) {
	poolCfg := gopool.Config{
		Constructor: newHTTPClient,
	}

	pool, err := gopool.NewPool(poolCfg)
	if err != nil {
		return nil, err
	}
	return &SimpleClient{
		pool: pool,
	}, nil
}

func (this *SimpleClient) Do(req *http.Request, result interface{}) (*http.Response, error) {
	c := this.get()
	defer this.release(c)

	resp, err := c.Do(req)
	if err != nil {
		return nil, newSimpleClientError(CodeConnectionError, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		// read response body
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp, newSimpleClientError(CodeReadResponseBodyError, err)
		}

		apierr := &APIError{}

		if err := json.Unmarshal(b, apierr); err != nil {
			apierr.ErrorMsg = string(b)
		}

		apierr.StatusCode = resp.StatusCode
		return resp, apierr
	}

	// no need to parse response body
	if result == nil {
		return resp, nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, newSimpleClientError(CodeReadResponseBodyError, err)
	}

	if err := json.Unmarshal(b, result); err != nil {
		return resp, newSimpleClientError(CodeJsonDecodeError, err)
	}

	return resp, nil
}

func (this *SimpleClient) get() *http.Client {
	v, _ := this.pool.Get()
	return v.(*http.Client)
}

func (this *SimpleClient) release(c *http.Client) {
	this.pool.Put(c)
}
