package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/yazgazan/bacom"
)

func listCmd(args []string) {
	c, err := parseListFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}

	versions, err := bacom.FindVersions(c.Dir, false, c.Constraints)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	for _, dirname := range versions {
		err = listTests(dirname, c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}
}

func listTests(dirname string, conf listConf) error {
	if !conf.Filenames {
		fmt.Printf("%s:\n", dirname)
	}

	reqFiles, err := bacom.GetRequestsFiles(dirname)
	if err != nil {
		return err
	}

	for _, fname := range reqFiles {
		req, err := parseRequest("", fname)
		if err != nil {
			return err
		}
		if conf.Filters.Match(req) != nil {
			continue
		}
		if conf.Filenames {
			fmt.Println(fname)
			continue
		}
		printTestDetails(conf, fname, req)
	}

	return nil
}

func printTestDetails(conf listConf, fname string, req *http.Request) {
	fmt.Printf("\t%s %s\n", req.Method, req.URL)
	if conf.Long {
		fmt.Printf("\t\tPath:                     %s\n", fname)
		if req.Method == http.MethodPost {
			cType := req.Header.Get("Content-Type")
			if cType != "" {
				fmt.Printf("\t\t(Request) Content-Type:   %s\n", cType)
			}
			fmt.Printf("\t\t(Request) Content-Length: %d\n", req.ContentLength)
		}
		resp, err := getBaseResponse(req, fname, targetConf{})
		if err != nil {
			fmt.Println("\t\t(response missing)")
			return
		}
		fmt.Printf("\t\tStatus:                   %s\n", resp.Status)
		cType := resp.Header.Get("Content-Type")
		if cType != "" {
			fmt.Printf("\t\tContent-Type:             %s\n", cType)
		}
		fmt.Printf("\t\tContent-Length:           %d\n", resp.ContentLength)
	}
}
