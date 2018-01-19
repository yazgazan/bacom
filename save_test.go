package backomp

import (
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/pkg/errors"
)

func TestSave(t *testing.T) {
	needClosing := make(map[string]io.Closer)
	defer func() {
		for fname, closer := range needClosing {
			err := closer.Close()
			if err != nil {
				t.Logf("failed to close %q: %s", fname, err)
			}
		}
	}()

	testSrcDir := path.Join(os.TempDir(), "testSrc")
	testDstDir := path.Join(os.TempDir(), "testDst")
	testReqFname := path.Join(testSrcDir, "foo_req.txt")
	// testRespFname := path.Join(testDstDir, "foo_resp.txt")

	req, err := http.NewRequest("GET", "http://example.org", nil)
	if err != nil {
		t.Fatalf("Error creating test request: %s", err)
	}

	err = os.Mkdir(testSrcDir, 0700)
	if err != nil {
		t.Fatalf("failed to create test src dir: %s", err)
	}
	defer func() {
		err = os.RemoveAll(testSrcDir)
		if err != nil {
			t.Logf("failed to remove test src dir: %s", err)
		}
	}()

	err = os.Mkdir(testDstDir, 0700)
	if err != nil {
		t.Errorf("failed to create test dst dir: %s", err)
		return
	}
	defer func() {
		err = os.RemoveAll(testDstDir)
		if err != nil {
			t.Logf("failed to remove test dst dir: %s", err)
		}
	}()

	f, err := os.Create(testReqFname)
	if err != nil {
		t.Errorf("failed to create test req file: %s", err)
		return
	}

	err = req.Write(f)
	if err != nil {
		t.Errorf("failed to write test req to file: %s", err)
		err = f.Close()
		if err != nil {
			t.Logf("failed to close test req file: %s", err)
		}
		return
	}

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: map[string][]string{
			"Content-Length": {"42"},
			"Content-Type":   {"application/json"},
		},
		Request: req,
	}

	err = Save(testDstDir, testReqFname, resp, map[string]interface{}{
		"foo": []interface{}{"bar"},
	})
	if err != nil {
		t.Errorf("Save(...): unexpected error: %s", err)
	}
}

func TestSaveFail(t *testing.T) {
	// Save should return an error if the req file is not properly formated
	testFile := path.Join(os.TempDir(), "foo.txt")
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create test file %q: %s", testFile, err)
	}
	defer func(testFile string) {
		err = os.Remove(testFile)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", testFile, err)
		}
	}(testFile)
	err = f.Close()
	if err != nil {
		t.Errorf("failed to close test file %q: %s", testFile, err)
		return
	}
	err = Save(os.TempDir(), testFile, &http.Response{}, nil)
	if err == nil {
		t.Errorf("Save(..., %q, ...): expected error, got nil", "foo.txt")
	}

	// Save should return an error if the req file doesn't exist
	err = Save(os.TempDir(), "foo_req.txt", &http.Response{}, nil)
	if err == nil {
		t.Errorf("Save(..., %q, ...): expected error, got nil", "foo.txt")
	}

	// Save should return an error if the body cannot be marshaled to JSON
	err = Save(os.TempDir(), "foo_req.txt", &http.Response{}, jsonFailer{})
	if err == nil {
		t.Errorf("Save(..., %q, ...): expected error, got nil", "foo.txt")
	}

	// Saving to a non-existant folder should result in an error
	testFile = path.Join(os.TempDir(), "foo_req.txt")
	f, err = os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create test file %q: %s", testFile, err)
	}
	defer func(testFile string) {
		err = os.Remove(testFile)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", testFile, err)
		}
	}(testFile)
	err = f.Close()
	if err != nil {
		t.Errorf("failed to close test file %q: %s", testFile, err)
		return
	}
	err = Save("does_not_exist", testFile, &http.Response{}, nil)
	if err == nil {
		t.Errorf("Save(%q, ...): expected error, got nil", "does_not_exist")
	}

	// Save should return an error if destination resp file cannot be created
	testDir := path.Join(os.TempDir(), "TestSaveFail")
	err = os.Mkdir(testDir, 0700)
	if err != nil {
		t.Errorf("failed to create test dir %q: %s", testDir, err)
		return
	}
	defer func(testDir string) {
		err = os.RemoveAll(testDir)
		if err != nil {
			t.Errorf("failed to remove test dir %q: %s", testDir, err)
		}
	}(testDir)
	err = os.Mkdir(path.Join(testDir, "bar_resp.txt"), 0700)
	if err != nil {
		t.Errorf("failed to create test dir %q: %s", path.Join(testDir, "bar_resp.txt"), err)
	}
	testFile = path.Join(os.TempDir(), "bar_req.txt")
	f, err = os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create test file %q: %s", testFile, err)
	}
	defer func(testFile string) {
		err = os.Remove(testFile)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", testFile, err)
		}
	}(testFile)
	err = f.Close()
	if err != nil {
		t.Errorf("failed to close test file %q: %s", testFile, err)
		return
	}
	err = Save(testDir, testFile, &http.Response{}, nil)
	if err == nil {
		t.Errorf("Save(%q, %q, ...): expected error, got nil", testDir, testFile)
	}
}

type jsonFailer struct{}

func (jsonFailer) MarshalJSON() ([]byte, error) {
	return nil, errors.New("test json error")
}
