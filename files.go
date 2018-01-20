package backomp

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	// ErrReqInvalidName is returned by GetResponseFilename when the filename
	// does not follow the _req[0-9]*.txt pattern
	ErrReqInvalidName = errors.New("invalid filename for request")
)

// GetRequestsFiles returns a list of request files matching the _req[0-9]*.txt pattern
func GetRequestsFiles(dirname string) (files []string, err error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, errors.Wrapf(err, "finding requests in %q", dirname)
	}
	defer handleClose(&err, f)

	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, errors.Wrapf(err, "finding requests in %q", dirname)
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		if !isRequestFilename(fi.Name()) {
			continue
		}

		files = append(files, filepath.Join(dirname, fi.Name()))
	}

	return files, nil
}

func isRequestFilename(fname string) bool {
	idx := strings.LastIndex(fname, "_req")
	if idx == -1 {
		return false
	}
	if !strings.HasSuffix(fname[idx:], ".txt") {
		return false
	}
	if idx+len("_req.txt") == len(fname) {
		return true
	}
	_, err := strconv.Atoi(fname[idx+len("_req") : len(fname)-len(".txt")])

	return err == nil
}

// GetResponseFilename transform a _req[0-9]*.txt filename into a _resp[0-9]*.txt
func GetResponseFilename(reqFname string) (string, error) {
	idx := strings.LastIndex(reqFname, "_req")
	if idx == -1 {
		return "", ErrReqInvalidName
	}
	if !strings.HasSuffix(reqFname[idx:], ".txt") {
		return "", ErrReqInvalidName
	}
	if idx+len("_req.txt") == len(reqFname) {
		return reqFname[0:idx] + "_resp.txt", nil
	}
	n, err := strconv.Atoi(reqFname[idx+len("_req") : len(reqFname)-len(".txt")])
	if err != nil {
		return "", ErrReqInvalidName
	}

	return reqFname[0:idx] + "_resp" + strconv.Itoa(n) + ".txt", nil
}
