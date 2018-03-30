package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yazgazan/bacom"

	"github.com/pkg/errors"
)

func cpCmd(args []string) {
	c, err := parseCpFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}

	err = cpFiles(c.Src, c.Dst)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func cpFiles(srcs []string, dst string) error {
	if len(srcs) > 1 {
		return cpFilesToDir(srcs, dst)
	}

	dstIsDir, err := dstInfo(dst)
	if err != nil {
		return err
	}
	if dstIsDir {
		return cpFileToDir(srcs[0], dst)
	}

	return cpFileToFile(srcs[0], dst)
}

func cpFilesToDir(srcs []string, dir string) error {
	dirFi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !dirFi.IsDir() {
		return errors.Errorf("%q is not a directory", dir)
	}

	for _, src := range srcs {
		err = cpFileToDir(src, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func cpFileToDir(src, dir string) error {
	name, err := bacom.NameFromReqFileName(src)
	if err != nil {
		return err
	}
	reqFname := bacom.ReqFileName(name, dir)
	err = cpFile(src, reqFname)
	if err != nil {
		return err
	}

	srcResp, err := bacom.GetResponseFilename(src)
	if err != nil {
		return err
	}
	dstResp, err := bacom.GetResponseFilename(reqFname)
	if err != nil {
		return err
	}
	respExists, err := fileExists(srcResp)
	if err != nil || !respExists {
		return err
	}

	return cpFile(srcResp, dstResp)
}

func cpFileToFile(src, dst string) error {
	dir := filepath.Dir(dst)
	name, err := bacom.NameFromReqFileName(dst)
	if err != nil {
		return err
	}
	reqFname := bacom.ReqFileName(name, dir)
	err = cpFile(src, reqFname)
	if err != nil {
		return err
	}

	srcResp, err := bacom.GetResponseFilename(src)
	if err != nil {
		return err
	}
	dstResp, err := bacom.GetResponseFilename(reqFname)
	if err != nil {
		return err
	}
	respExists, err := fileExists(srcResp)
	if err != nil || !respExists {
		return err
	}

	return cpFile(srcResp, dstResp)
}

func cpFile(srcFname, dstFname string) (err error) {
	src, err := os.Open(srcFname)
	if err != nil {
		return err
	}
	defer handleClose(&err, src)

	dst, err := os.Create(dstFname)
	if err != nil {
		return err
	}
	defer handleClose(&err, dst)

	_, err = io.Copy(dst, src)

	if err != nil {
		return err
	}
	fmt.Printf("%s -> %s\n", srcFname, dstFname)
	return nil
}
