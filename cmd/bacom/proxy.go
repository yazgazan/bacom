package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func importProxyCmd(args []string) {
	c, err := parseImportProxyFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}

	err = os.MkdirAll(c.Dir, 0750)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatal(err)
	}

	err = runProxy(c.Listen, c.Target, c.Dir, c.Verbose, c.Filters)

	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatal(err)
	}
}

func runProxy(listen, target, outDir string, verbose bool, filters reqFilters) error {
	targetURL, err := url.Parse(target)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    listen,
		Handler: proxyHandler(targetURL, outDir, verbose, filters),
	}

	log.Printf("listening on %s", listen)
	return srv.ListenAndServe()
}

func proxyHandler(target *url.URL, outDir string, verbose bool, filters reqFilters) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusInternalServerError)
			log.Printf("failed to read request body: %v", err)
			return
		}

		u := *target // copies the URL

		u.Path = path.Join(u.Path, r.URL.Path)
		q := u.Query()
		for k, v := range r.URL.Query() {
			q[k] = v
		}
		u.RawQuery = q.Encode()
		if target.Fragment == "" {
			u.Fragment = r.URL.Fragment
		}

		fmt.Println(u.String())

		req, err := http.NewRequest(r.Method, u.String(), bytes.NewReader(reqBody))
		if err != nil {
			http.Error(w, "failed to create proxied request", http.StatusInternalServerError)
			log.Printf("failed to create proxied request: %v", err)
			return
		}
		for k, vv := range r.Header {
			if k == "Host" || k == "Transfer-Encoding" || k == "Accept-Encoding" {
				continue
			}

			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "failed to proxy request", http.StatusBadGateway)
			log.Printf("failed to proxy request: %v", err)
			return
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "failed to read response body", http.StatusInternalServerError)
			log.Printf("failed to read response body: %v", err)
			return
		}

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, bytes.NewReader(respBody))
		if err != nil {
			log.Printf("failed to write response: %v", err)
			return
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
		if err := filters.Match(req); err != nil {
			if verbose {
				log.Printf("skipping %s: %v", u.String(), err)
			}
			return
		}

		name := strings.ToLower(req.Method) + "-" + normalize(u.Path)

		reqFname, err := importReq(verbose, outDir, name, req)
		if err != nil {
			if verbose {
				log.Printf("failed to save request for %q: %v", u.String(), err)
			}
			return
		}

		resp.Body = ioutil.NopCloser(bytes.NewReader(respBody))
		err = importResp(verbose, reqFname, outDir, name, resp)
		if err != nil {
			if verbose {
				log.Printf("failed to save response for %q: %v", u.String(), err)
			}
			return
		}

		log.Printf("saved req/resp %q for %q", name, u.String())
	}
}
