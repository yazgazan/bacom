package bacom

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveResponse(t *testing.T) {
	needClosing := make(map[string]io.Closer)
	defer func() {
		for fname, closer := range needClosing {
			err := closer.Close()
			if err != nil {
				t.Logf("failed to close %q: %s", fname, err)
			}
		}
	}()

	testSrcDir := filepath.Join(os.TempDir(), "testSrc")
	testDstDir := filepath.Join(os.TempDir(), "testDst")
	testReqFname := filepath.Join(testSrcDir, "foo_req.txt")

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
	err = f.Close()
	if err != nil {
		t.Logf("failed to close test req file: %s", err)
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

	err = SaveResponse(testDstDir, testReqFname, resp)
	if err != nil {
		t.Errorf("SaveResponse(...): unexpected error: %s", err)
	}
}

func TestSaveResponseFail(t *testing.T) {
	// SaveResponse should return an error if the req file is not properly formated
	testFile := filepath.Join(os.TempDir(), "foo.txt")
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
	err = SaveResponse(os.TempDir(), testFile, &http.Response{})
	if err == nil {
		t.Errorf("SaveResponse(..., %q, ...): expected error, got nil", "foo.txt")
	}

	// Saving to a non-existent folder should result in an error
	testFile = filepath.Join(os.TempDir(), "foo_req.txt")
	err = SaveResponse("does_not_exist", testFile, &http.Response{})
	if err == nil {
		t.Errorf("SaveResponse(%q, ...): expected error, got nil", "does_not_exist")
	}

	// SaveResponse should return an error if destination resp file cannot be created
	testDir := filepath.Join(os.TempDir(), "TestSaveResponseFail")
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
	err = os.Mkdir(filepath.Join(testDir, "bar_resp.txt"), 0700)
	if err != nil {
		t.Errorf("failed to create test dir %q: %s", filepath.Join(testDir, "bar_resp.txt"), err)
	}
	testFile = filepath.Join(os.TempDir(), "bar_req.txt")
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
	err = SaveResponse(testDir, testFile, &http.Response{})
	if err == nil {
		t.Errorf("SaveResponse(%q, %q, ...): expected error, got nil", testDir, testFile)
	}
}

func TestSaveRequest(t *testing.T) {
	fname := filepath.Join(os.TempDir(), "foo_req.txt")
	testDstDir := filepath.Join(os.TempDir(), "testDst")

	// SaveRequest should fail if the source file doesn't exist
	err := SaveRequest(testDstDir, fname)
	if err == nil {
		t.Errorf("SaveRequest(%q, %q): expected error, got nil", testDstDir, fname)
	}

	f, err := os.Create(fname)
	if err != nil {
		t.Fatalf("failed to create test file %q: %s", fname, err)
	}
	defer func(fname string) {
		err = os.Remove(fname)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", fname, err)
		}
	}(fname)
	err = f.Close()
	if err != nil {
		t.Errorf("failed to close test file %q: %s", fname, err)
		return
	}

	// SaveRequest should fail if the destination folder doesn't exist
	folderNotExist := filepath.Join(os.TempDir(), "DoesNotExist")
	err = SaveRequest(folderNotExist, fname)
	if err == nil {
		t.Errorf("SaveRequest(%q, %q): expected error, got nil", folderNotExist, fname)
	}

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

	err = SaveRequest(testDstDir, fname)
	if err != nil {
		t.Errorf("SaveRequest(%q, %q): unexpected error: %s", testDstDir, fname, err)
	}
}
