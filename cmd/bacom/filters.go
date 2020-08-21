package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/yazgazan/bacom"
)

type reqFilters struct {
	Paths         stringsFlag
	IgnorePaths   stringsFlag
	Hosts         regexesFlag
	IgnoreHosts   regexesFlag
	Methods       stringsFlag
	IgnoreMethods stringsFlag
	ReqBody       stringsFlag
	IgnoreReqBody stringsFlag
}

func (f *reqFilters) SetupFlags(flags *flag.FlagSet) {
	flags.Var(&f.Paths, "paths", "path patterns to import (can be repeated)")
	flags.Var(&f.IgnorePaths, "ignore-paths", "path patterns to ignore (can be repeated)")
	flags.Var(&f.Hosts, "hosts", "host regexes to import (can be repeated)")
	flags.Var(&f.IgnoreHosts, "ignore-hosts", "host regexes to ignore (can be repeated)")
	flags.Var(&f.Methods, "methods", "methods to import (can be repeated)")
	flags.Var(&f.IgnoreMethods, "ignore-methods", "methods to ignore (can be repeated)")
	flags.Var(&f.ReqBody, "req-body", "include if request body contains (can be repeated)")
	flags.Var(&f.IgnoreReqBody, "ignore-req-body", "exclude if request body contains (can be repeated)")
}

func (f reqFilters) Match(req *http.Request) error {

	err := f.matchMethod(req)
	if err != nil {
		return err
	}

	err = f.matchPath(req)
	if err != nil {
		return err
	}

	err = f.matchBody(req)
	if err != nil {
		return err
	}

	err = f.matchHost(req)

	return err
}

func (f reqFilters) matchBody(req *http.Request) error {
	if len(f.ReqBody) == 0 && len(f.IgnoreReqBody) == 0 {
		return nil
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	err = req.Body.Close()
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	err = matchBodyContains(req, body, f.ReqBody...)
	if err != nil {
		return err
	}

	err = matchBodyExclude(req, body, f.IgnoreReqBody...)

	return err
}

func matchBodyContains(req *http.Request, body []byte, ss ...string) error {
	for _, s := range ss {
		if !bytes.Contains(body, []byte(s)) {
			return fmt.Errorf("body missing %q", s)
		}
	}

	return nil
}

func matchBodyExclude(req *http.Request, body []byte, ss ...string) error {
	for _, s := range ss {
		if bytes.Contains(body, []byte(s)) {
			return fmt.Errorf("body contains %q", s)
		}
	}

	return nil
}

func (f reqFilters) matchPath(req *http.Request) error {
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

func (f reqFilters) matchHost(req *http.Request) error {
	hostname := req.URL.Hostname()
	if hostname == "" {
		hostname = req.Host
	}

	foundMatch := len(f.Hosts) == 0
	for _, h := range f.Hosts {
		if h.MatchString(hostname) {
			foundMatch = true
		}
	}
	if !foundMatch {
		return errors.Errorf("%s %q does not match the host filters", req.Method, req.URL)
	}
	for _, h := range f.IgnoreHosts {
		if h.MatchString(hostname) {
			return errors.Errorf("%s %q does not match the host filters", req.Method, req.URL)
		}
	}

	return nil
}

func (f reqFilters) matchMethod(req *http.Request) error {
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
