package apple

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// public const
const (
	ProtocolVersion = "QH65B2"
	UserLocale      = "en_US"
	ClientID        = "XABBG36SBA"
	APIKey          = "ba2ec180e6ca6e6c6a542255453b24d6e6e5b2be0cc48bc1b0d8ad64cfe0228f"
)

type clientRequest struct {
	url    string
	req    *http.Request
	header map[string]string
	resp   *http.Response
	err    error
}

// RequestHeader header
var RequestHeader map[string]string = map[string]string{
	"Host":            "developerservices2.apple.com",
	"Accept":          "text/x-xml-plist",
	"Content-Type":    "text/x-xml-plist",
	"Accept-Language": "en-us",
	"Accept-Encoding": "gzip, deflate",
	"X-Xcode-Version": "7.0 (7A120f)",
}

// JSONRequestHeader header
var JSONRequestHeader map[string]string = map[string]string{
	"Accept":           "application/vnd.api+json",
	"Content-Type":     "application/vnd.api+json",
	"X-Apple-App-Info": "com.apple.gs.xcode.auth",
	"X-Xcode-Version":  "7.0 (7A120f)",
	"Accept-Encoding":  "gzip, deflate",
	"User-Agent":       "Xcode",
}

// NewClientRequest init
func NewClientRequest(rawurl, method string) *clientRequest {
	var resp http.Response
	req, err := http.NewRequest(method, rawurl, nil)
	if err != nil {
		panic("create request error")
	}
	return &clientRequest{rawurl, req, map[string]string{}, &resp, err}
}

// SetRawURL function
func (c *clientRequest) SetRawURL(rawurl string) *clientRequest {
	c.req.URL, c.err = url.Parse(rawurl)
	if c.err == nil {
		c.req.Host = c.req.URL.Host
		c.req.Header = make(http.Header)
		c.req.Body = nil
		c.req.ContentLength = 0
	}
	return c
}

// SetMethod function
func (c *clientRequest) SetMethod(method string) *clientRequest {
	c.req.Method = method
	return c
}

// SetBody function
func (c *clientRequest) SetBody(data interface{}) *clientRequest {
	switch t := data.(type) {
	case string:
		buffer := bytes.NewBufferString(t)
		c.req.Body = ioutil.NopCloser(buffer)
		c.req.ContentLength = int64(len(t))
	case []byte:
		buffer := bytes.NewBuffer(t)
		c.req.Body = ioutil.NopCloser(buffer)
		c.req.ContentLength = int64(len(t))
	}
	return c
}

// SetHeader header
func (c *clientRequest) SetHeader(header map[string]string) *clientRequest {
	for k, v := range header {
		c.req.Header.Set(k, v)
	}
	return c
}

// GetResponse infomation
func (c *clientRequest) GetResponse() (*http.Response, error) {
	client := &http.Client{}
	c.resp, c.err = client.Do(c.req)
	if c.err != nil {
		return nil, c.err
	}
	return c.resp, nil
}

// GetBody get response body
func (c *clientRequest) GetBody() ([]byte, int, error) {
	resp, err := c.GetResponse()
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var reader io.Reader
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, 0, err
		}
	} else {
		reader = resp.Body
	}
	var body []byte
	statusCode := resp.StatusCode
	body, c.err = ioutil.ReadAll(reader)
	return body, statusCode, nil
}
