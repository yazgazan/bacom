package bacom

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// SaveRequest generates the _req.txt file corresponding to provided filename
func SaveRequest(dir, fname string) error {
	reqName := filepath.Base(fname)

	return copyFile(fname, filepath.Join(dir, reqName))
}

// SaveResponse generates the _resp.txt file corresponding to reqFname and the provided response/body
func SaveResponse(dir, reqFname string, resp *http.Response) (err error) {
	reqName := filepath.Base(reqFname)
	respName, err := GetResponseFilename(reqName)
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
