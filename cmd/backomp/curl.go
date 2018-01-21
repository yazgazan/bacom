package main

import (
	"github.com/yazgazan/backomp"

	"fmt"
	"os"
	"net/http"
	"io"
	"bytes"
	"path/filepath"
)

func curlCmd(args []string) {
	c, err := parseCurlFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	defer func() {
		err = c.Data.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}()

	buf := &bytes.Buffer{}
	r := io.TeeReader(c.Data, buf)
	req, err := newRequest(c.Method, c.URL, http.Header(c.Headers), r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if c.Name == "" {
		resp.Write(os.Stdout)
		return
	}

	req, err = newRequest(c.Method, c.URL, http.Header(c.Headers), buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	reqFile := filepath.Join(c.Dir, c.Name + "_req.txt")
	f, err := os.Create(reqFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	err = req.Write(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	respFile, err := backomp.GetResponseFilename(reqFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	f, err = os.Create(respFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	defer f.Close()
	err = resp.Write(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

}

func newRequest(method, url string, h http.Header, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header = h

	return req, nil
}
