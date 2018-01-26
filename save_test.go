package bacom

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
)

func createTestFolder(t *testing.T, dirname string) (remover func(), err error) {
	err = os.MkdirAll(dirname, 0700)
	if err != nil {
		return nil, err
	}
	return func() {
		err = os.RemoveAll(dirname)
		if err != nil {
			t.Errorf("failed to remove test dir %q: %s", dirname, err)
		}
	}, nil
}

func createTestRequest(fname string, body []byte) error {
	req, err := http.NewRequest("POST", "http://example.org", bytes.NewBuffer(body))
	if err != nil {
		return errors.Wrap(err, "failed to create test request")
	}
	f, err := os.Create(fname)
	if err != nil {
		return errors.Wrapf(err, "failed to open test src file %q", fname)
	}
	defer f.Close()
	err = req.Write(f)
	if err != nil {
		return errors.Wrapf(err, "failed to write test request to %q", fname)
	}

	return nil
}

func TestSaveRequest(t *testing.T) {
	testDirV0 := filepath.Join(os.TempDir(), "testDirV0")
	removeTestDirV0, err := createTestFolder(t, testDirV0)
	if err != nil {
		t.Fatalf("failed to create test dir %q: %s", testDirV0, err)
	}
	defer removeTestDirV0()

	testDirV1 := filepath.Join(os.TempDir(), "testDirV1")
	removeTestDirV1, err := createTestFolder(t, testDirV1)
	if err != nil {
		t.Errorf("failed to create test dir %q: %s", testDirV1, err)
		return
	}
	defer removeTestDirV1()

	testSrc := filepath.Join(testDirV0, "foo_req.txt")
	saver := NewSaver(testDirV1, testSrc)

	// SaveRequest should fail if the source file does not exist
	err = saver.SaveRequest()
	if err == nil {
		t.Errorf("saving from a non-existent file should fail.")
	}

	createTestRequest(testSrc, []byte(`{"foo": ["bar"]}`))

	// SaveRequest should succeed if the destination file does not exist
	testDst := filepath.Join(testDirV1, "foo_req.txt")
	err = saver.SaveRequest()
	if err != nil {
		t.Errorf("failed to save to non-existent file %q: %s", testDst, err)
	}

	// SaveRequest should succeed if the destination file is identical to the source
	err = saver.SaveRequest()
	if err != nil {
		t.Errorf("failed to save to identical file %q: %s", testDst, err)
	}

	// SaveRequest should succeed if the destination file is different from the source
	createTestRequest(testSrc, []byte(`{"foo": {"bar": 42}}`))

	err = saver.SaveRequest()
	if err != nil {
		t.Errorf("failed to save to different file %q: %s", testDst, err)
	}
	testDst1 := filepath.Join(testDirV1, "foo_req1.txt")

	if !fileExists(testDst1) {
		t.Errorf("expected %q to exists", testDst1)
	}
}

func TestSaveRequestFail(t *testing.T) {
	// ErrReqInvalidName should be returned when the req filename is mal-formated
	testFile := "foo.txt"
	saver := NewSaver(os.TempDir(), testFile)
	err := saver.SaveRequest()
	if err != ErrReqInvalidName {
		t.Errorf("Saver{%q}.SaveRequest(): expected error %q, got %v", testFile, ErrReqInvalidName, err)
	}

	// SaveRequest should fail if the source file does not exist
	testFile = filepath.Join(os.TempDir(), "non-existent_req.txt")
	saver = NewSaver(os.TempDir(), testFile)
	err = saver.SaveRequest()
	if err == nil {
		t.Errorf("Saver{%q}.SaveRequest(): expected error, got nil instead", testFile)
	}

	testFile = filepath.Join(os.TempDir(), "foo_req.txt")
	f, err := os.Create(testFile)
	if err != nil {
		t.Errorf("error creating test file %q: %s", testFile, err)
		return
	}
	defer func(testFile string) {
		err = os.Remove(testFile)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", testFile, err)
		}
	}(testFile)
	err = f.Close()
	if err != nil {
		t.Errorf("error closing test file %q: %s", testFile, err)
		return
	}

	testSrcFile := filepath.Join(os.TempDir(), "non-existent", "foo_req.txt")
	saver = NewSaver(os.TempDir(), testSrcFile)
	err = saver.SaveRequest()
	if err == nil {
		t.Errorf("Saver{%q}.SaveRequest(): expected error, got nil instead", testSrcFile)
	}
}

func TestSaveResponse(t *testing.T) {
	testFile := filepath.Join(os.TempDir(), "foo_resp.txt")
	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", testFile, err)
		}
	}()
	saver := NewSaver(os.TempDir(), "foo_req.txt")
	err := saver.SaveResponse(&http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: ioutil.NopCloser(bytes.NewBufferString(`{"foo": "bar"}`)),
	})
	if err != nil {
		t.Errorf("failed to save response: %s", err)
	}
}

func TestSaveResponseFail(t *testing.T) {
	// foo.txt is not a valid request filename
	saver := NewSaver(os.TempDir(), "foo.txt")
	err := saver.SaveResponse(&http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: ioutil.NopCloser(bytes.NewBufferString(`{"foo": "bar"}`)),
	})
	if err == nil {
		t.Error("expected error saving response using invalid file name")
	}

	saver = NewSaver(filepath.Join(os.TempDir(), "non-existent"), "foo_req.txt")
	err = saver.SaveResponse(&http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: ioutil.NopCloser(bytes.NewBufferString(`{"foo": "bar"}`)),
	})
	if err == nil {
		t.Error("expected error saving response to non-existent folder")
	}
}

func validateResponseFile(fname string) (err error) {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer handleClose(&err, f)

	_, err = http.ReadResponse(bufio.NewReader(f), nil)

	return err
}

func TestCompareFiles(t *testing.T) {
	testLHSFile := filepath.Join(os.TempDir(), "lhsfile.txt")
	testRHSFile := filepath.Join(os.TempDir(), "rhsfile.txt")
	err := ioutil.WriteFile(testLHSFile, []byte(`foo`), 0600)
	if err != nil {
		t.Errorf("failed to write test file %q: %s", testLHSFile, err)
		return
	}
	defer func(fname string) {
		err = os.Remove(fname)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", fname, err)
		}
	}(testLHSFile)
	err = ioutil.WriteFile(testRHSFile, []byte(`foo`), 0600)
	if err != nil {
		t.Errorf("failed to write test file %q: %s", testRHSFile, err)
		return
	}
	defer func(fname string) {
		err = os.Remove(fname)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", fname, err)
		}
	}(testRHSFile)

	match, err := compareFiles(testLHSFile, testRHSFile)
	if err != nil {
		t.Errorf("compareFiles(%q, %q): unexpected error: %s", testLHSFile, testRHSFile, err)
	}
	if !match {
		t.Errorf("compareFiles(%q, %q) = %v, expected %v", testLHSFile, testRHSFile, match, true)
	}

	err = ioutil.WriteFile(testLHSFile, []byte(`bar`), 0600)
	if err != nil {
		t.Errorf("failed to write test file %q: %s", testLHSFile, err)
		return
	}
	match, err = compareFiles(testLHSFile, testRHSFile)
	if err != nil {
		t.Errorf("compareFiles(%q, %q): unexpected error: %s", testLHSFile, testRHSFile, err)
	}
	if match {
		t.Errorf("compareFiles(%q, %q) = %v, expected %v", testLHSFile, testRHSFile, match, false)
	}

	testNonExistent := filepath.Join(os.TempDir(), "file-not-exist.txt")
	_, err = compareFiles(testLHSFile, testNonExistent)
	if err == nil {
		t.Errorf("compareFiles(%q, %q): expected error, got nil", testLHSFile, testNonExistent)
	}

	_, err = compareFiles(testNonExistent, testRHSFile)
	if err == nil {
		t.Errorf("compareFiles(%q, %q): expected error, got nil", testNonExistent, testRHSFile)
	}
}

func TestCopyFile(t *testing.T) {
	testSrcFile := filepath.Join(os.TempDir(), "srcFile.txt")
	testDstFile := filepath.Join(os.TempDir(), "dstFile.txt")
	err := ioutil.WriteFile(testSrcFile, []byte(`foo`), 0600)
	if err != nil {
		t.Errorf("failed to write test file %q: %s", testSrcFile, err)
		return
	}
	defer func(fname string) {
		err = os.Remove(fname)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", fname, err)
		}
	}(testSrcFile)
	defer func(fname string) {
		err = os.Remove(fname)
		if err != nil {
			t.Errorf("failed to remove test file %q: %s", fname, err)
		}
	}(testDstFile)

	err = copyFile(testSrcFile, testDstFile)
	if err != nil {
		t.Errorf("copyFile(%q, %q): unexpected error: %s", testSrcFile, testDstFile, err)
	}

	testNonExistent := filepath.Join(os.TempDir(), "non-existent", "folder-not-exist.txt")
	err = copyFile(testSrcFile, testNonExistent)
	if err == nil {
		t.Errorf("copyFile(%q, %q): expected error, got nil", testSrcFile, testNonExistent)
	}
}

func TestCompareReaders(t *testing.T) {
	b0, err := randomBytes(5000)
	if err != nil {
		t.Error("getting random bytes:", err)
	}
	erroringB0 := ErroringReader{bytes.NewReader(b0)}
	b1, err := randomBytes(5000)
	if err != nil {
		t.Error("getting random bytes:", err)
	}
	b2, err := randomBytes(100)
	if err != nil {
		t.Error("getting random bytes:", err)
	}

	for _, test := range []struct {
		name     string
		lhs, rhs io.Reader
		expected bool
	}{
		{"b0, b0", bytes.NewReader(b0), bytes.NewReader(b0), true},
		{"b0, b1", bytes.NewReader(b0), bytes.NewReader(b1), false},
		{"b0, b2", bytes.NewReader(b0), bytes.NewReader(b2), false},
		{"b0, b0[:100]", bytes.NewReader(b0), bytes.NewReader(b0[:100]), false},
		{"b0[:100], b0", bytes.NewReader(b0[:100]), bytes.NewReader(b0), false},
		{"b0[:100], b0[:100]", bytes.NewReader(b0[:100]), bytes.NewReader(b0[:100]), true},
		{"b0.b2.b1, b0.b2.b1", bytes.NewReader(cat(b0, b2, b1)), bytes.NewReader(cat(b0, b2, b1)), true},
		{"b0.b2.b1, b0.b2.b0", bytes.NewReader(cat(b0, b2, b1)), bytes.NewReader(cat(b0, b2, b0)), false},
		{"[], []", bytes.NewReader(nil), bytes.NewReader(nil), true},
		{"[], b0", bytes.NewReader(nil), bytes.NewReader(b0), false},
		{"b0, []", bytes.NewReader(b0), bytes.NewReader(nil), false},
	} {
		v, err := compareReaders(test.lhs, test.rhs)
		if err != nil {
			t.Errorf("compareReaders(%s): unexpected error: %s", test.name, err)
		}
		if v != test.expected {
			t.Errorf("compareReaders(%s) = %v, expected %v", test.name, v, test.expected)
		}
	}

	_, err = compareReaders(bytes.NewReader(b0), erroringB0)
	if err == nil {
		t.Error("compareReaders(b0, erroringB0): expected error, got nil")
	}

	_, err = compareReaders(erroringB0, bytes.NewReader(b0))
	if err == nil {
		t.Error("compareReaders(erroringB0, b0): expected error, got nil")
	}
}

type ErroringReader struct {
	io.Reader
}

func (r ErroringReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err != nil && err != io.EOF {
		return n, err
	}

	return n, errors.New("some error")
}

func cat(bufs ...[]byte) []byte {
	var b []byte

	for _, buf := range bufs {
		b = append(b, buf...)
	}

	return b
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	return b, err
}
