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
		err = listTests(dirname, c.Long, c.Filters)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}
}

func listTests(dirname string, long bool, filters reqFilters) error {
	fmt.Printf("%s:\n", dirname)

	reqFiles, err := bacom.GetRequestsFiles(dirname)
	if err != nil {
		return err
	}

	for _, fname := range reqFiles {
		req, err := parseRequest("", fname)
		if err != nil {
			return err
		}
		if filters.Match(req) != nil {
			continue
		}
		fmt.Printf("\t%s %s\n", req.Method, req.URL)
		if long {
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
