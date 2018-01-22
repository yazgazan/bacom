package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/yazgazan/backomp"
)

func testCmd(args []string) {
	c, err := parseTestFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	versions, err := backomp.FindVersions(c.Dir, c.Verbose, c.Constraints)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if c.Save != "" {
		if err = os.MkdirAll(filepath.Join(c.Dir, c.Save), 0700); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}

	failed := false
	for _, dirname := range versions {
		pass, err := runTestsForVersion(c, dirname)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		printPass(pass, c.Quiet, dirname)
		failed = failed || !pass
	}

	if failed {
		os.Exit(1)
	}
}

func printPass(pass, quiet bool, fname string) {
	if pass && !quiet {
		fmt.Println("OK  ", fname)
	}
	if !pass {
		fmt.Println("FAIL", fname)
	}
}

func runTestsForVersion(conf testConf, dirname string) (bool, error) {
	reqFiles, err := backomp.GetRequestsFiles(dirname)
	if err != nil {
		return false, errors.Wrapf(err, "looking for requests files in %q", dirname)
	}

	passed := true
	for _, fname := range reqFiles {
		ok, err := runTest(conf, fname)
		if err != nil {
			return false, err
		}
		printPass(ok, conf.Quiet, filepath.Join(dirname, fname))
		passed = passed && ok
	}

	return passed, nil
}

func runTest(conf testConf, fname string) (pass bool, err error) {
	targetResp, baseResp, reqPath, err := getResponses(conf, fname)
	if targetResp != nil {
		defer handleClose(&err, targetResp.Body)
	}
	if baseResp != nil {
		defer handleClose(&err, baseResp.Body)
	}
	if err != nil {
		return false, errors.Wrapf(err, "getting responses for %q", fname)
	}

	pConf := getPathConf(conf.Paths, reqPath)

	results, err := backomp.CompareHeaders(
		pConf.Headers.Ignore,
		pConf.Headers.IgnoreContent,
		baseResp.Header,
		targetResp.Header,
	)
	if err != nil {
		return false, errors.Wrapf(err, "comparing headers for %q", fname)
	}

	targetBody, err := readBody(targetResp)
	if err != nil {
		return false, errors.Wrapf(err, "reading target response body")
	}
	baseBody, err := readBody(baseResp)
	if err != nil {
		return false, errors.Wrapf(err, "reading base response body")
	}
	bodyResults, err := backomp.Compare(
		pConf.JSON.Ignore,
		pConf.JSON.IgnoreMissing,
		pConf.JSON.IgnoreNull,
		baseBody,
		targetBody,
	)
	if err != nil {
		return false, errors.Wrapf(err, "comparing bodies")
	}
	if conf.Save != "" {
		err = backomp.Save(filepath.Join(conf.Dir, conf.Save), fname, targetResp, targetBody)
		if err != nil {
			return false, errors.Wrapf(err, "saving request/response to %s for %q", conf.Save, fname)
		}
	}
	results = append(results, bodyResults...)

	printResults(fname, results)

	return len(results) == 0, nil
}

func printResults(fname string, results []string) {
	if len(results) != 0 {
		fmt.Printf("\n%s:\n", fname)
	}
	for _, result := range results {
		fmt.Println(result)
	}
}

func readBody(resp *http.Response) (body interface{}, err error) {
	err = json.NewDecoder(resp.Body).Decode(&body)

	return body, err
}

func getResponses(conf testConf, fname string) (target, base *http.Response, path string, err error) {
	req, err := parseRequest(fname)
	if err != nil {
		return nil, nil, "", err
	}
	target, err = getTargetResponse(req, fname, conf.Target)
	if err != nil {
		return nil, nil, "", errors.Wrapf(err, "getting target response for %q", fname)
	}

	req, err = parseRequest(fname)
	if err != nil {
		return target, nil, "", err
	}
	base, err = getBaseResponse(req, fname, conf.Base)
	if err != nil {
		return target, nil, "", errors.Wrapf(err, "getting base response for %q", fname)
	}

	return target, base, req.URL.Path, nil
}

func parseRequest(fname string) (req *http.Request, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing request %q", fname)
	}
	defer handleClose(&err, f)

	req, err = http.ReadRequest(bufio.NewReader(f))

	return req, errors.Wrapf(err, "parsing request %q", fname)
}

func getTargetResponse(req *http.Request, reqFname string, targetConf targetConf) (*http.Response, error) {
	return getTargetResponseFromHost(req, targetConf.Host, targetConf.UseHTTPS)
}

func getBaseResponse(req *http.Request, reqFname string, targetConf targetConf) (*http.Response, error) {
	if targetConf.Host != "" {
		return getTargetResponseFromHost(req, targetConf.Host, targetConf.UseHTTPS)
	}

	return backomp.ReadResponse(req, reqFname)
}

func getTargetResponseFromHost(req *http.Request, host string, useHTTPS bool) (*http.Response, error) {
	req.Host = host
	req.RequestURI = ""
	req.URL.Host = host
	if useHTTPS {
		req.URL.Scheme = "https"
	} else {
		req.URL.Scheme = "http"
	}

	return http.DefaultClient.Do(req)
}
