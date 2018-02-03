package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type stringSlice []string

func (ss stringSlice) String() string {
	return strings.Join(ss, ",")
}

func (ss *stringSlice) Set(s string) error {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	if len(parts) > 1 {
		for _, part := range parts {
			err := ss.Set(part)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if contains(*ss, s) {
		return nil
	}

	*ss = append(*ss, s)
	return nil
}

type headers http.Header

func (h headers) String() string {
	return fmt.Sprintf("%q", http.Header(h))
}

func (h headers) Set(s string) error {
	if s == "" {
		return nil
	}

	if !strings.Contains(s, ":") {
		return fmt.Errorf("marlformated header %q", s)
	}

	parts := strings.SplitN(s, ":", 2)
	http.Header(h).Set(parts[0], parts[1])

	return nil
}

func contains(ss []string, needle string) bool {
	for _, s := range ss {
		if s == needle {
			return true
		}
	}

	return false
}

func main() {
	var r io.ReadCloser
	var err error

	var removeHeaders stringSlice
	setHeaders := make(headers)

	flag.Var(&removeHeaders, "remove-header", "header to remove (can be repeated)")
	flag.Var(&setHeaders, "set-header", "header to set (Key: Value, can be repeated)")
	flag.Parse()

	r, err = preprocessReq(os.Stdin, removeHeaders, setHeaders)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = r.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatal(err)
	}
}

func preprocessReq(in io.Reader, removeHeaders []string, setHeaders headers) (io.ReadCloser, error) {
	req, err := http.ReadRequest(bufio.NewReader(in))
	if err != nil {
		return nil, err
	}

	for _, h := range removeHeaders {
		req.Header.Del(h)
	}

	for name, values := range setHeaders {
		for _, value := range values {
			req.Header.Set(name, value)
		}
	}

	b := bytes.NewBuffer(nil)
	err = req.Write(b)
	return ioutil.NopCloser(b), err
}
