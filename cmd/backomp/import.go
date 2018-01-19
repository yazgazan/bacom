package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/yazgazan/backomp/har"
)

func importCmd(args []string) {
	c, err := parseImportFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	for _, fname := range c.Files {
		err := importFromFile(fname, c.Dir)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func normalize(s string) string {
	replacer := strings.NewReplacer("/", "-", "-", "--")

	return replacer.Replace(s)
}

func reqFileName(name, dest string) (string, error) {
	return _fileName(name, dest, "_req", 0)
}

func respFileName(name, dest string) (string, error) {
	return _fileName(name, dest, "_resp", 0)
}

func _fileName(name, dest, suffix string, i int) (string, error) {
	var fname string

	if i == 0 {
		fname = path.Join(dest, name+suffix+".txt")
	} else {
		fname = path.Join(dest, name+fmt.Sprintf("%s%d.txt", suffix, i))
	}

	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		return fname, nil
	}
	if err != nil {
		return "", err
	}

	return _fileName(name, dest, suffix, i+1)
}

func importFromFile(fname, outDir string) (err error) {
	var harObj har.HAR

	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer handleClose(&err, f)

	err = json.NewDecoder(f).Decode(&harObj)
	if err != nil {
		return err
	}

	for _, entry := range harObj.Log.Entries {
		u, err := url.Parse(entry.Request.URL)
		if err != nil {
			return err
		}
		req, err := entry.Request.ToHTTPRequest(u.Host, false)
		if err != nil {
			return err
		}
		resp, err := entry.Response.ToHTTPResponse(req)
		if err != nil {
			return err
		}
		name := normalize(u.Path)

		err = importReq(outDir, name, req)
		if err != nil {
			return err
		}

		err = importResp(outDir, name, resp)
		if err != nil {
			return err
		}
	}

	return nil
}

func importReq(outDir, name string, req *http.Request) (err error) {
	fname, err := reqFileName(name, outDir)
	if err != nil {
		return err
	}
	outF, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer handleClose(&err, outF)

	return req.Write(outF)
}

func importResp(outDir, name string, resp *http.Response) (err error) {
	fname, err := respFileName(name, outDir)
	if err != nil {
		return err
	}
	outF, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer handleClose(&err, outF)

	return resp.Write(outF)
}

func handleClose(err *error, closer io.Closer) {
	errClose := closer.Close()
	if *err == nil {
		*err = errClose
	}
}
