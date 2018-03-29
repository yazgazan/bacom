package bacom

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/Masterminds/semver"
)

func TestFindVersions(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "TestFindVersions")
	testDirs := []string{
		filepath.Join(testDir, "v."),
		filepath.Join(testDir, "v0.0.1"),
		filepath.Join(testDir, "v0.0.3"),
		filepath.Join(testDir, "v0.1.0"),
		filepath.Join(testDir, "v0.1.6"),
		filepath.Join(testDir, "v1.1.0"),
		filepath.Join(testDir, "v2.0.0"),
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
	expected := []string{filepath.Join(testDir, "v0.1.0")}
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
		filepath.Join(testDir, "v0.1.0"),
		filepath.Join(testDir, "v0.1.6"),
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

func TestVersionMatch(t *testing.T) {
	for _, test := range []struct {
		constraint string
		input      string
		expected   bool
	}{
		{
			constraint: "v0.2.1",
			input:      "v0.2.1",
			expected:   true,
		},
		{
			constraint: "v0.2.1",
			input:      "v0.2.3",
			expected:   false,
		},
		{
			constraint: "v1.x",
			input:      "v1.0.4",
			expected:   true,
		},
		{
			constraint: "v1.x",
			input:      "v2.0.4",
			expected:   false,
		},
	} {
		constraints, err := semver.NewConstraint(test.constraint)
		if err != nil {
			t.Errorf("failed to parse constraint %q: %s", test.constraint, err)
			continue
		}

		valid, err := VersionMatch(true, constraints, test.input)
		if err != nil {
			t.Errorf("VersionMatch(true, %q, %q): unexpected error: %s", test.constraint, test.input, err)
			continue
		}
		if valid != test.expected {
			t.Errorf("VersionMatch(true, %q, %q) = %v, expected %v", test.constraint, test.input, valid, test.expected)
		}
	}
}

func TestVersionMatchFail(t *testing.T) {
	for _, test := range []struct {
		constraint string
		input      string
	}{
		{
			constraint: "v0.2.1",
			input:      "v",
		},
		{
			constraint: "v0.2.1",
			input:      "",
		},
		{
			constraint: "v1.x",
			input:      "v1.a.2",
		},
		{
			constraint: "v1.x",
			input:      "v2.",
		},
	} {
		constraints, err := semver.NewConstraint(test.constraint)
		if err != nil {
			t.Errorf("failed to parse constraint %q: %s", test.constraint, err)
			continue
		}

		_, err = VersionMatch(true, constraints, test.input)
		if err == nil {
			t.Errorf("VersionMatch(true, %q, %q): expected error, got nil", test.constraint, test.input)
			continue
		}
	}
}
