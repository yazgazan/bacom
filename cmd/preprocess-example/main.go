package main

import (
	"bufio"
	"bytes"
	"flag"
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

	var isReq bool
	var isResp bool
	var removeHeaders stringSlice

	flag.BoolVar(&isReq, "req", false, "parsing request")
	flag.BoolVar(&isResp, "resp", false, "parsing response")
	flag.Var(&removeHeaders, "remove-header", "header to remove (can be repeated)")
	flag.Parse()

	if isReq && isResp {
		log.Fatal("conflicting -req and -resp")
	}
	if !isReq && !isResp {
		log.Fatal("need one of -req or -resp")
	}

	if isReq {
		r, err = preprocessReq(os.Stdin, removeHeaders)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		r, err = preprocessResp(os.Stdin, removeHeaders)
		if err != nil {
			log.Fatal(err)
		}
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

func preprocessReq(in io.Reader, removeHeaders []string) (io.ReadCloser, error) {
	req, err := http.ReadRequest(bufio.NewReader(in))
	if err != nil {
		return nil, err
	}

	for _, h := range removeHeaders {
		req.Header.Del(h)
	}

	b := bytes.NewBuffer(nil)
	err = req.Write(b)
	return ioutil.NopCloser(b), err
}

func preprocessResp(in io.Reader, removeHeaders []string) (io.ReadCloser, error) {
	req, err := http.ReadResponse(bufio.NewReader(in), nil)
	if err != nil {
		return nil, err
	}

	for _, h := range removeHeaders {
		req.Header.Del(h)
	}

	b := bytes.NewBuffer(nil)
	err = req.Write(b)
	return ioutil.NopCloser(b), err
}
