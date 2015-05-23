package jsonutils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tukdesk/gopool"
)

const (
	CodeConnectionError = iota
	CodeReadResponseBodyError
	CodeJsonDecodeError
)

func newHTTPClient() (interface{}, error) {
	return &http.Client{}, nil
}

func newClientError(code int, err error) error {
	return NewAPIError(0, code, err.Error())
}

type Client struct {
	pool *gopool.Pool
}

func NewClient() (*Client, error) {
	poolCfg := gopool.Config{
		Constructor: newHTTPClient,
	}

	pool, err := gopool.NewPool(poolCfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		pool: pool,
	}, nil
}

func (this *Client) Do(req *http.Request, result interface{}) error {
	c := this.get()
	defer this.release(c)

	resp, err := c.Do(req)
	if err != nil {
		return newClientError(CodeConnectionError, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		// read response body
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return newClientError(CodeReadResponseBodyError, err)
		}

		apierr := &APIError{}

		if err := json.Unmarshal(b, apierr); err != nil {
			apierr.ErrorMsg = string(b)
		}

		apierr.StatusCode = resp.StatusCode
		return apierr
	}

	// no need to parse response body
	if result == nil {
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return newClientError(CodeReadResponseBodyError, err)
	}

	if err := json.Unmarshal(b, result); err != nil {
		return newClientError(CodeJsonDecodeError, err)
	}

	return nil
}

func (this *Client) get() *http.Client {
	v, _ := this.pool.Get()
	return v.(*http.Client)
}

func (this *Client) release(c *http.Client) {
	this.pool.Put(c)
}
