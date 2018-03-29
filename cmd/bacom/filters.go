package main

import (
	"fmt"
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

	err = f.matchHost(req)

	return err
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
