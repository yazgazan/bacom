package main

import (
	"fmt"
	"os"

	"github.com/yazgazan/bacom"
)

func listCmd(args []string) {
	c, err := parseListFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
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
		fmt.Printf("\t%s %s\n", req.Method, req.URL)
		if conf.Long {
			fmt.Printf("\t\tPath:           %s\n", fname)
			resp, err := getBaseResponse(req, fname, targetConf{})
			if err != nil {
				fmt.Println("\t\t(response missing)")
				continue
			}
			fmt.Printf("\t\tStatus:         %s\n", resp.Status)
			fmt.Printf("\t\tContent-Length: %d\n", resp.ContentLength)
		}
	}

	return nil
}
