package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/yazgazan/bacom"
	"github.com/yazgazan/bacom/har"
)

func importHarCmd(args []string) {
	c, err := parseImportHARFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}

	for _, fname := range c.Files {
		err := importFromFile(fname, c.Dir, c.Verbose, c.Filters)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func normalize(s string) string {
	var out []byte

	for _, c := range []byte(s) {
		if c != '/' {
			out = append(out, c)
			continue
		}
		if len(out) == 0 {
			continue
		}
		if out[len(out)-1] == '-' {
			continue
		}
		out = append(out, '-')
	}

	if len(out) != 0 && out[len(out)-1] == '-' {
		return string(out[:len(out)-1])
	}

	return string(out)
}

func importFromFile(fname, outDir string, verbose bool, filters reqFilters) (err error) {
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
		u, req, resp, err := parseEntry(entry)
		if err != nil {
			return err
		}

		err = filters.Match(req)
		if err != nil && verbose {
			fmt.Fprintf(os.Stderr, "excluding request: %s\n", err)
		}
		if err != nil {
			continue
		}

		name := strings.ToLower(req.Method) + "-" + normalize(u.Path)

		reqFname, err := importReq(verbose, outDir, name, req)
		if err != nil {
			return err
		}

		err = importResp(verbose, reqFname, outDir, name, resp)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseEntry(entry har.Entry) (*url.URL, *http.Request, *http.Response, error) {
	u, err := url.Parse(entry.Request.URL)
	if err != nil {
		return nil, nil, nil, err
	}
	req, err := entry.Request.ToHTTPRequest(u.Host, false)
	if err != nil {
		return nil, nil, nil, err
	}

	resp, err := entry.Response.ToHTTPResponse(req)

	return u, req, resp, err
}

func importReq(verbose bool, outDir, name string, req *http.Request) (fname string, err error) {
	fname = bacom.ReqFileName(name, outDir)
	outF, err := os.Create(fname)
	if err != nil {
		return fname, err
	}
	defer handleClose(&err, outF)

	err = req.Write(outF)
	if err != nil {
		return fname, err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "imported %q\n", fname)
	}
	return fname, nil
}

func importResp(verbose bool, reqFname, outDir, name string, resp *http.Response) (err error) {
	fname, err := bacom.GetResponseFilename(reqFname)
	if err != nil {
		return err
	}
	outF, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer handleClose(&err, outF)

	if verbose {
		fmt.Fprintf(os.Stderr, "imported %q\n", fname)
	}
	return resp.Write(outF)
}

func handleClose(err *error, closer io.Closer) {
	errClose := closer.Close()
	if *err == nil {
		*err = errClose
	}
}
