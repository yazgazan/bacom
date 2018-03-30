package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yazgazan/bacom"

	"github.com/pkg/errors"
)

func mvCmd(args []string) {
	c, err := parseMvFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}

	err = mvFiles(c.Src, c.Dst)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func mvFiles(srcs []string, dst string) error {
	if len(srcs) > 1 {
		return mvFilesToDir(srcs, dst)
	}

	dstIsDir, err := dstInfo(dst)
	if err != nil {
		return err
	}
	if dstIsDir {
		return mvFileToDir(srcs[0], dst)
	}

	return mvFileToFile(srcs[0], dst)
}

func mvFilesToDir(srcs []string, dir string) error {
	dirFi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !dirFi.IsDir() {
		return errors.Errorf("%q is not a directory", dir)
	}

	for _, src := range srcs {
		err = mvFileToDir(src, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func mvFileToDir(src, dir string) error {
	name, err := bacom.NameFromReqFileName(src)
	if err != nil {
		return err
	}
	reqFname := bacom.ReqFileName(name, dir)
	err = mvFile(src, reqFname)
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

	return mvFile(srcResp, dstResp)
}

func mvFileToFile(src, dst string) error {
	dir := filepath.Dir(dst)
	name, err := bacom.NameFromReqFileName(dst)
	if err != nil {
		return err
	}
	reqFname := bacom.ReqFileName(name, dir)
	err = mvFile(src, reqFname)
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

	return mvFile(srcResp, dstResp)
}

func mvFile(srcFname, dstFname string) (err error) {
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

	fmt.Printf("%s -> %s\n", srcFname, dstFname)
	return err
}

func dstInfo(fname string) (isDir bool, err error) {
	fi, err := os.Stat(fname)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return fi.IsDir(), nil
}

func fileExists(fname string) (bool, error) {
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
