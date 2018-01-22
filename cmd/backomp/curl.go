package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yazgazan/backomp"
)

func closeOrExit(c io.Closer) {
	err := c.Close()
	logAndExitOnError(err)
}

func logAndExitOnError(err error) {
	if err == nil {
		return
	}

	log.Print("Error:", err)
	os.Exit(1)
}

func curlCmd(args []string) {
	c, err := parseCurlFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	defer closeOrExit(c.Data)

	buf := &bytes.Buffer{}
	r := io.TeeReader(c.Data, buf)
	req, err := newRequest(c.Method, c.URL, http.Header(c.Headers), r)
	logAndExitOnError(err)

	if c.Name == "" {
		err = req.Write(os.Stdout)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	logAndExitOnError(err)

	req, err = newRequest(c.Method, c.URL, http.Header(c.Headers), buf)
	logAndExitOnError(err)

	reqFile := filepath.Join(c.Dir, c.Name+"_req.txt")
	f, err := os.Create(reqFile)
	logAndExitOnError(err)
	err = req.Write(f)
	logAndExitOnError(err)

	respFile, err := backomp.GetResponseFilename(reqFile)
	logAndExitOnError(err)
	f, err = os.Create(respFile)
	logAndExitOnError(err)
	defer closeOrExit(f)
	err = resp.Write(f)
	logAndExitOnError(err)

}

func newRequest(method, url string, h http.Header, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header = h

	return req, nil
}
