package backomp

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestIsRequestFilename(t *testing.T) {
	for _, test := range []struct {
		In    string
		Value bool
	}{
		{"_req.txt", true},
		{"", false},
		{"foo_req.txt", true},
		{"foo_req1.txt", true},
		{"foo_req2.txt", true},
		{"foo_req42.txt", true},
		{"bar_REQ.txt", false},
		{"foo.txt", false},
		{"_req.txt.txt", false},
		{"_req.txt_req.txt", true},
	} {
		v := isRequestFilename(test.In)
		if v != test.Value {
			t.Errorf("isRequestFilename(%q) = %v, expected %v", test.In, v, test.Value)
		}
	}
}

func TestGetResponseFilename(t *testing.T) {
	for _, test := range []struct {
		In       string
		Expected string
		Err      error
	}{
		{"_req.txt", "_resp.txt", nil},
		{"foo_req.txt", "foo_resp.txt", nil},
		{"foo_req1.txt", "foo_resp1.txt", nil},
		{"foo_req2.txt", "foo_resp2.txt", nil},
		{"foo_req24.txt", "foo_resp24.txt", nil},
		{"_req.txt_req.txt", "_req.txt_resp.txt", nil},
		{"foo.txt", "", ErrReqInvalidName},
		{"foo_req.go", "", ErrReqInvalidName},
		{"_req.txt.txt", "", ErrReqInvalidName},
		{"", "", ErrReqInvalidName},
	} {
		v, err := GetResponseFilename(test.In)
		if v != test.Expected {
			t.Errorf("getResponseFilename(%q) = %q, expected %q", test.In, v, test.Expected)
		}
		if err == nil && test.Err != nil {
			t.Errorf("getResponseFilename(%q): got nil, expected error %q", test.In, test.Err)
		}
		if err != nil && test.Err == nil {
			t.Errorf("getResponseFilename(%q): unexpected error %q", test.In, err)
		}
	}
}

func TestGetRequestsFiles(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "TestGetRequestsFiles")
	testFiles := []string{
		filepath.Join(testDir, "foo.txt"),
		filepath.Join(testDir, "foo_req.go"),
		filepath.Join(testDir, "foo_resp.txt"),
		filepath.Join(testDir, "foo_req.txt"),
		filepath.Join(testDir, "foo_req0.txt"),
		filepath.Join(testDir, "foo_req1.txt"),
		filepath.Join(testDir, "foo_req2.txt"),
		filepath.Join(testDir, "foo_req10.txt"),
	}
	expectedFiles := []string{
		filepath.Join(testDir, "foo_req.txt"),
		filepath.Join(testDir, "foo_req0.txt"),
		filepath.Join(testDir, "foo_req1.txt"),
		filepath.Join(testDir, "foo_req2.txt"),
		filepath.Join(testDir, "foo_req10.txt"),
	}
	sort.Strings(expectedFiles)

	err := os.Mkdir(testDir, 0700)
	if err != nil {
		t.Fatalf("failed to create test dir: %s", testDir)
	}
	defer func() {
		err = os.RemoveAll(testDir)
		if err != nil {
			t.Logf("failed to remove test dir: %s", err)
		}
	}()

	for _, fname := range testFiles {
		f, err := os.Create(fname)
		if err != nil {
			t.Errorf("failed to create test file %q: %s", fname, err)
			return
		}
		err = f.Close()
		if err != nil {
			t.Errorf("failed to close test file %q: %s", fname, err)
			return
		}
	}

	err = os.Mkdir(filepath.Join(testDir, "foo_req11.txt"), 0700)
	if err != nil {
		t.Errorf("failed to create test dir: %s", err)
	}

	files, err := GetRequestsFiles(testDir)
	if err != nil {
		t.Errorf("GetRequestsFiles(%q): unexpected error: %s", testDir, err)
	}

	sort.Strings(files)
	if !reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("GetRequestsFiles(%q): %q, expected %q", testDir, files, expectedFiles)
	}
}

func TestGetRequestsFilesFail(t *testing.T) {
	nonExistant := filepath.Join(os.TempDir(), "does_not_exist")
	_, err := GetRequestsFiles(nonExistant)
	if err == nil {
		t.Errorf("GetRequestsFiles(%q): expected error, got nil", nonExistant)
	}

	fileNotDir := filepath.Join(os.TempDir(), "not_a_dir")
	f, err := os.Create(fileNotDir)
	if err != nil {
		t.Errorf("failed to create test file %q: %s", fileNotDir, err)
		return
	}
	defer func() {
		err = os.Remove(fileNotDir)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", fileNotDir, err)
		}
	}()
	err = f.Close()
	if err != nil {
		t.Errorf("failed to close test file %q: %s", fileNotDir, err)
		return
	}

	_, err = GetRequestsFiles(fileNotDir)
	if err == nil {
		t.Errorf("GetRequestsFiles(%q): expected error, got nil", fileNotDir)
	}
}
