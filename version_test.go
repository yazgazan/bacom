package backomp

import (
	"os"
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/Masterminds/semver"
)

func TestFindVersions(t *testing.T) {
	testDir := path.Join(os.TempDir(), "TestFindVersions")
	testDirs := []string{
		path.Join(testDir, "v."),
		path.Join(testDir, "v0.0.1"),
		path.Join(testDir, "v0.0.3"),
		path.Join(testDir, "v0.1.0"),
		path.Join(testDir, "v0.1.6"),
		path.Join(testDir, "v1.1.0"),
		path.Join(testDir, "v2.0.0"),
	}

	newConstraint := func(s string) *semver.Constraints {
		c, err := semver.NewConstraint(s)
		if err != nil {
			t.Errorf("failed to create constraint %q: %s", s, err)
		}

		return c
	}

	err := os.Mkdir(testDir, 0700)
	if err != nil {
		t.Fatalf("failed to create test dir %q: %s", testDir, err)
	}
	defer func() {
		err = os.RemoveAll(testDir)
		if err != nil {
			t.Errorf("failed to remove test dir %q: %s", testDir, err)
		}
	}()

	for _, fname := range testDirs {
		err = os.Mkdir(fname, 0700)
		if err != nil {
			t.Errorf("failed to create test dir %q: %s", fname, err)
			return
		}
	}

	cStr := "0.1.0"
	c := newConstraint(cStr)
	if c == nil {
		return
	}
	files, err := FindVersions(testDir, true, c)
	expected := []string{path.Join(testDir, "v0.1.0")}
	if err != nil {
		t.Errorf("FindVersions(%q, false, %q): unexpected error: %s", testDir, cStr, err)
	}
	if !reflect.DeepEqual(files, expected) {
		t.Errorf("FindVersions(%q, false, %q) = %q, expected %q", testDir, cStr, files, expected)
	}

	cStr = "0.1.x"
	c = newConstraint(cStr)
	if c == nil {
		return
	}
	files, err = FindVersions(testDir, true, c)
	sort.Strings(files)
	expected = []string{
		path.Join(testDir, "v0.1.0"),
		path.Join(testDir, "v0.1.6"),
	}
	sort.Strings(expected)
	if err != nil {
		t.Errorf("FindVersions(%q, false, %q): unexpected error: %s", testDir, cStr, err)
	}
	if !reflect.DeepEqual(files, expected) {
		t.Errorf("FindVersions(%q, false, %q) = %q, expected %q", testDir, cStr, files, expected)
	}
}

func TestFindVersionsFail(t *testing.T) {
	_, err := FindVersions("does_not_exist", false, nil)
	if err == nil {
		t.Errorf("FindVersions(%q): expected error, got nil", "does_not_exist")
	}

	fileNotDir := path.Join(os.TempDir(), "not_a_dir")
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

	_, err = FindVersions(fileNotDir, false, nil)
	if err == nil {
		t.Errorf("FindVersions(%q): expected error, got nil", fileNotDir)
	}

	cStr := "99.99.99"
	c, err := semver.NewConstraint(cStr)
	if err != nil {
		t.Errorf("failed to create constraint %q: %s", cStr, err)
		return
	}
	_, err = FindVersions(os.TempDir(), false, c)
	if err == nil {
		t.Errorf("FindVersions(%q, false, %q): expected error, got nil", os.TempDir(), cStr)
	}
}
