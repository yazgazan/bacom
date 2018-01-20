package backomp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// Save generates the _req.txt and _resp.txt files corresponding to reqFname and the provided response/body
func Save(dir, reqFname string, resp *http.Response, body interface{}) (err error) {
	reqName := filepath.Base(reqFname)
	respName, err := GetResponseFilename(reqName)
	if err != nil {
		return err
	}

	b := &bytes.Buffer{}
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return err
	}
	resp.Body = ioutil.NopCloser(b)
	resp.ContentLength = int64(b.Len())
	if _, ok := resp.Header["Content-Length"]; ok {
		resp.Header.Set("Content-Length", strconv.Itoa(b.Len()))
	}

	err = copyFile(reqFname, filepath.Join(dir, reqName))
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, respName))
	if err != nil {
		return err
	}
	defer handleClose(&err, f)

	return resp.Write(f)
}

func copyFile(srcFname, dstFname string) (err error) {
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

	return err
}

func handleClose(err *error, closer io.Closer) {
	errClose := closer.Close()
	if *err == nil {
		*err = errClose
	}
}
