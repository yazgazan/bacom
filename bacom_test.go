package bacom

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestReadResponse(t *testing.T) {
	testFname := filepath.Join(os.TempDir(), "foo_resp.txt")
	testReqFname := filepath.Join(os.TempDir(), "foo_req.txt")
	testContent := []byte(`{
		"foo": ["bar"]
	}`)
	testContentBuff := bytes.NewBuffer(testContent)

	req, err := http.NewRequest("GET", "http://example.org", nil)
	if err != nil {
		t.Fatalf("Error creating test request: %s", err)
	}

	_, err = ReadResponse(req, "foo.txt")
	if err == nil {
		t.Errorf("ReadResponse(req, %q): expected error, got nil", "foo.txt")
	}

	_, err = ReadResponse(req, "foo_req.txt")
	if err == nil {
		t.Errorf("ReadResponse(req, %q): expected error, got nil", "foo_req.txt")
	}

	f, err := os.Create(testFname)
	if err != nil {
		t.Fatalf("Error creating test response file: %s", err)
	}
	defer func() {
		err = os.Remove(testFname)
		if err != nil {
			t.Logf("error removing test response file: %s", err)
		}
	}()

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: map[string][]string{
			"Content-Length": {strconv.Itoa(testContentBuff.Len())},
			"Content-Type":   {"application/json"},
		},
		Body:          ioutil.NopCloser(testContentBuff),
		ContentLength: int64(testContentBuff.Len()),
		Request:       req,
	}

	err = resp.Write(f)
	if err != nil {
		errClose := f.Close()
		if errClose != nil {
			t.Logf("error closing test response file: %s", errClose)
		}
		t.Errorf("writing test response: %s", err)
		return
	}
	err = f.Close()
	if err != nil {
		t.Logf("error closing test response file: %s", err)
		return
	}

	result, err := ReadResponse(req, testReqFname)
	if err != nil {
		t.Errorf("ReadResponse(req, %q): unexpected error: %s", testReqFname, err)
	}
	resp.Body = nil
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Errorf("failed to read result's body: %s", err)
	}
	err = result.Body.Close()
	result.Body = nil
	if err != nil {
		t.Errorf("failed to close result's body: %s", err)
	}

	if !reflect.DeepEqual(result, resp) {
		t.Errorf("ReadResponse(req, %q) = %+v, expected %+v", testReqFname, result, resp)
	}
	if bytes.Compare(body, testContent) != 0 {
		t.Errorf("ReadResponse(req, %q): body does not match", testReqFname)
	}
}
