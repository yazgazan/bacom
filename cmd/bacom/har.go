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

	"github.com/pkg/errors"

	"github.com/yazgazan/bacom"
	"github.com/yazgazan/bacom/har"
)

func importHarCmd(args []string) {
	c, err := parseImportHARFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	for _, fname := range c.Files {
		err := importFromFile(fname, c.Dir, c.Verbose, c.Filters)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type harFilters struct {
	Paths         stringsFlag
	IgnorePaths   stringsFlag
	Hosts         regexesFlag
	IgnoreHosts   regexesFlag
	Methods       stringsFlag
	IgnoreMethods stringsFlag
}

func (f harFilters) Match(req *http.Request) error {

	err := f.matchMethod(req)
	if err != nil {
		return err
	}

	err = f.matchPath(req)
	if err != nil {
		return err
	}

	err = f.matchHost(req)

	return err
}

func (f harFilters) matchPath(req *http.Request) error {
	foundMatch := len(f.Paths) == 0
	for _, p := range f.Paths {
		match, err := bacom.MatchPath(p, req.URL.Path)
		if err != nil {
			return err
		}
		if match {
			foundMatch = true
			break
		}
	}
	if !foundMatch {
		return errors.Errorf("%s %q does not match the path filters", req.Method, req.URL)
	}
	for _, p := range f.IgnorePaths {
		match, err := bacom.MatchPath(p, req.URL.Path)
		if err != nil {
			fmt.Printf("%s %q excluded by error: %s\n", req.Method, req.URL, err)
			return err
		}
		if match {
			return errors.Errorf("%s %q does not match the path filters", req.Method, req.URL)
		}
	}

	return nil
}

func (f harFilters) matchHost(req *http.Request) error {
	foundMatch := len(f.Hosts) == 0
	for _, h := range f.Hosts {
		if h.MatchString(req.URL.Hostname()) {
			foundMatch = true
		}
	}
	if !foundMatch {
		return errors.Errorf("%s %q does not match the host filters", req.Method, req.URL)
	}
	for _, h := range f.IgnoreHosts {
		if h.MatchString(req.URL.Hostname()) {
			return errors.Errorf("%s %q does not match the host filters", req.Method, req.URL)
		}
	}

	return nil
}

func (f harFilters) matchMethod(req *http.Request) error {
	foundMatch := len(f.Methods) == 0
	for _, m := range f.Methods {
		if strings.ToLower(req.Method) == strings.ToLower(m) {
			foundMatch = true
		}
	}
	if !foundMatch {
		return errors.Errorf("%s %q does not match the method filters", req.Method, req.URL)
	}
	for _, m := range f.IgnoreMethods {
		if strings.ToLower(req.Method) == strings.ToLower(m) {
			return errors.Errorf("%s %q does not match the method filters", req.Method, req.URL)
		}
	}

	return nil
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

func importFromFile(fname, outDir string, verbose bool, filters harFilters) (err error) {
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
