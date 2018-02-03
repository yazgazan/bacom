package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/yazgazan/bacom"
	"golang.org/x/sync/errgroup"
)

func testCmd(args []string) {
	c, err := parseTestFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	versions, err := bacom.FindVersions(c.Dir, c.Verbose, c.Constraints)
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
	reqFiles, err := bacom.GetRequestsFiles(dirname)
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

func compareStatuses(lhsCode, rhsCode int, lhs, rhs string) []string {
	if lhsCode == rhsCode {
		return nil
	}

	return []string{
		"- (Status) " + color.RedString(lhs),
		"+ (Status) " + color.GreenString(rhs),
	}
}

func compareResponses(
	conf testConf,
	reqPath, reqMethod,
	fname string,
	baseResp, targetResp *http.Response,
) (results []string, err error) {
	pConf := getPathConf(conf.Paths, reqMethod, reqPath)

	targetBody, err := readBody(targetResp)
	if err != nil {
		return nil, errors.Wrapf(err, "reading target response body")
	}
	baseBody, err := readBody(baseResp)
	if err != nil {
		return nil, errors.Wrapf(err, "reading base response body")
	}

	results, err = bacom.CompareHeaders(
		pConf.Headers.Ignore,
		pConf.Headers.IgnoreContent,
		baseResp.Header,
		targetResp.Header,
	)
	if err != nil {
		return results, errors.Wrapf(err, "comparing headers for %q", fname)
	}

	results = append(compareStatuses(
		baseResp.StatusCode, targetResp.StatusCode,
		baseResp.Status, targetResp.Status,
	), results...)

	bodyResults, err := bacom.Compare(
		pConf.JSON.Ignore,
		pConf.JSON.IgnoreMissing,
		pConf.JSON.IgnoreNull,
		baseBody,
		targetBody,
	)
	if err != nil {
		return results, errors.Wrapf(err, "comparing bodies")
	}

	results = append(results, bodyResults...)

	return results, nil
}

type readCloser struct {
	io.Reader
	io.Closer
}

func runTest(conf testConf, fname string) (pass bool, err error) {
	var results []string

	targetResp, baseResp, reqPath, reqMethod, err := getResponses(conf, fname)
	if targetResp != nil {
		defer handleClose(&err, targetResp.Body)
	}
	if baseResp != nil {
		defer handleClose(&err, baseResp.Body)
	}
	if err != nil {
		return false, errors.Wrapf(err, "getting responses for %q", fname)
	}

	errg := &errgroup.Group{}

	if conf.Save != "" {
		b := &bytes.Buffer{}
		r := io.TeeReader(targetResp.Body, b)
		saveResp := &http.Response{}
		*saveResp = *targetResp
		targetResp.Body = readCloser{
			Reader: r,
			Closer: targetResp.Body,
		}
		saveResp.Body = ioutil.NopCloser(b)
		errg.Go(func() error {
			saver := bacom.NewSaver(filepath.Join(conf.Dir, conf.Save), fname)

			err = saver.SaveRequest()
			if err != nil {
				return err
			}

			err = saver.SaveResponse(saveResp)
			if err != nil {
				return errors.Wrapf(err, "saving request/response to %s for %q", conf.Save, fname)
			}

			return nil
		})
	}

	if baseResp != nil {
		errg.Go(func() error {
			var errCmp error

			results, errCmp = compareResponses(
				conf, reqPath, reqMethod, fname,
				baseResp, targetResp,
			)

			return errCmp
		})
	}

	err = errg.Wait()
	if err != nil {
		return false, err
	}

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
	if resp == nil {
		return nil, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&body)

	if err == io.EOF {
		return nil, nil
	}
	return body, err
}

func getResponses(conf testConf, fname string) (target, base *http.Response, path, method string, err error) {
	req, err := parseRequest(conf.Target.PreProcess, fname)
	if err != nil {
		return nil, nil, "", "", err
	}
	target, err = getTargetResponse(req, fname, conf.Target)
	if err != nil {
		return nil, nil, "", "", errors.Wrapf(err, "getting target response for %q", fname)
	}

	req, err = parseRequest(conf.Base.PreProcess, fname)
	if err != nil {
		return target, nil, "", "", err
	}
	base, err = getBaseResponse(req, fname, conf.Base)
	if os.IsNotExist(err) {
		return target, nil, req.URL.Path, req.Method, nil
	}
	if err != nil {
		return target, nil, "", "", errors.Wrapf(err, "getting base response for %q", fname)
	}

	return target, base, req.URL.Path, req.Method, nil
}

func parseRequest(preprocess, fname string) (req *http.Request, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing request %q", fname)
	}
	defer handleClose(&err, f)

	if preprocess == "" {
		req, err = http.ReadRequest(bufio.NewReader(f))

		return req, errors.Wrapf(err, "parsing request %q", fname)
	}

	// TODO(yazgazan): add timeout using the context
	cmd := exec.CommandContext(context.Background(), "/bin/sh", "-c", preprocess)
	cmd.Stdin = f
	b := &bytes.Buffer{}
	cmd.Stdout = b

	err = cmd.Run()
	if err != nil {
		return req, errors.Wrapf(err, "parsing request %q", fname)
	}

	req, err = http.ReadRequest(bufio.NewReader(b))

	return req, errors.Wrapf(err, "parsing request %q", fname)
}

func getTargetResponse(req *http.Request, reqFname string, targetConf targetConf) (*http.Response, error) {
	return getTargetResponseFromHost(req, targetConf.Host, targetConf.UseHTTPS)
}

func getBaseResponse(req *http.Request, reqFname string, targetConf targetConf) (*http.Response, error) {
	if targetConf.Host != "" {
		return getTargetResponseFromHost(req, targetConf.Host, targetConf.UseHTTPS)
	}

	return bacom.ReadResponse(req, reqFname)
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
