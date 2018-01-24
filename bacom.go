package bacom

import (
	"bufio"
	"net/http"
	"os"
)

// ReadResponse reads the response in reqFname given the provided request
func ReadResponse(req *http.Request, reqFname string) (resp *http.Response, err error) {
	fname, err := GetResponseFilename(reqFname)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(f), req)
}
