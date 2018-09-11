package har

import (
	"bytes"
	"io"
	"net/http"
)

// Response is the HAR representation of an http response
type Response struct {
	Status       int
	StatusText   string
	HTTPVersion  string
	RedirectURL  string
	HeadersSize  int64
	BodySize     int64
	TransferSize int64 `json:"_transferSize"`
	Headers      []KV
	Cookies      []Cookie
	Content      Content `json:"content"`
}

func (r *Response) getBody() (io.ReadCloser, error) {
	if len(r.Content.Text) == 0 {
		return nil, nil
	}
	return &readCloserWrapper{bytes.NewBuffer([]byte(r.Content.Text))}, nil
}

// ToHTTPResponse generate an *http.Response from a HAR response
func (r *Response) ToHTTPResponse(req *http.Request) (*http.Response, error) {
	headers := make(http.Header)
	for _, header := range r.Headers {
		if hName := http.CanonicalHeaderKey(header.Name); hName != "" {
			switch hName {
			case http.CanonicalHeaderKey("Cookie"):
				continue
			case http.CanonicalHeaderKey("Accept-Encoding"):
				continue
			}

			headers.Set(hName, header.Value)
		}
	}
	body, err := r.getBody()

	resp := &http.Response{
		Status:        r.StatusText,
		StatusCode:    r.Status,
		Proto:         "1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        headers,
		Body:          body,
		ContentLength: int64(len(r.Content.Text)),
		Request:       req,
	}

	return resp, err
}
