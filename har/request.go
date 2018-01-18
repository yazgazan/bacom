package har

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

// Request is the HAR representation of an http request
type Request struct {
	Method      string
	URL         string
	HTTPVersion string
	HeaderSize  int64
	BodySize    int64
	Headers     []KV
	QueryString []KV
	Cookies     []Cookie
	PostData    Content
	Content     Content `json:"content"`
}

// Content represents a request's or response's body
type Content struct {
	Size        int64
	MimeType    string
	Text        string
	Compression int
}

// KV is used to store key-value pairs for headers and queries
type KV struct {
	Name  string
	Value string
}

// Cookie represents a request's cookie
type Cookie struct {
	Name     string
	Value    string
	HTTPOnly bool
	Secure   bool
}

type readCloserWrapper struct {
	io.Reader
}

func (rc *readCloserWrapper) Close() error {
	return nil
}

func (r *Request) getBody() (io.ReadCloser, error) {
	if len(r.PostData.Text) == 0 {
		return nil, nil
	}
	return &readCloserWrapper{bytes.NewBuffer([]byte(r.PostData.Text))}, nil
}

// ToHTTPRequest generate an *http.Request from a HAR request
func (r *Request) ToHTTPRequest(host string, ForceHTTP bool) (*http.Request, error) {
	reqURL, err := url.Parse(r.URL)
	if ForceHTTP {
		reqURL.Scheme = "http"
	}
	if host != "" {
		reqURL.Host = host
	}
	if err != nil {
		return nil, err
	}
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

	return &http.Request{
		Method:        r.Method,
		URL:           reqURL,
		Header:        headers,
		Body:          body,
		GetBody:       r.getBody,
		ContentLength: int64(len(r.PostData.Text)),
	}, err
}
