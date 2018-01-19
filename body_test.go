package backomp

import (
	"strings"
	"testing"
)

func TestCompare(t *testing.T) {
	report, err := Compare(nil, nil, false, nil, nil)
	if err != nil {
		t.Errorf("Compare(nil, nil, false, nil, nil): unexpected error: %s", err)
	}
	if len(report) != 0 {
		t.Errorf(
			"Compare(nil, nil, false, nil, nil): unexpected report:\n%s\n",
			strings.Join(report, "\n"),
		)
	}

	mapA := map[string]interface{}{
		"foo":  "bar",
		"fizz": []interface{}{42},
		"buzz": 84,
	}
	mapB := map[string]interface{}{
		"foo":  "bar",
		"fizz": 23,
	}
	mapC := map[string]interface{}{
		"foo":  "bar",
		"fizz": 23,
		"buzz": nil,
	}
	recMap := map[string]interface{}{}
	recMap["foo"] = recMap

	report, err = Compare(nil, nil, false, mapA, mapA)
	if err != nil {
		t.Errorf("Compare(nil, nil, false, mapA, mapA): unexpected error: %s", err)
	}
	if len(report) != 0 {
		t.Errorf(
			"Compare(nil, nil, false, mapA, mapA): unexpected report:\n%s\n",
			strings.Join(report, "\n"),
		)
	}

	report, err = Compare(nil, nil, false, mapA, mapB)
	if err != nil {
		t.Errorf("Compare(nil, nil, false, mapA, mapB): unexpected error: %s", err)
	}
	if len(report) != 2 {
		t.Errorf(
			"Compare(nil, nil, false, mapA, mapB): unexpected report length (should be 2 lines): %d",
			len(report),
		)
	}

	// TODO(yazgazan): test ignoreNull
	report, err = Compare([]string{".fizz"}, nil, false, mapA, mapB)
	if err != nil {
		t.Errorf("Compare(%q, nil, false, mapA, mapB): unexpected error: %s", []string{".fizz"}, err)
	}
	if len(report) != 1 {
		t.Errorf(
			"Compare(%q, nil, false, mapA, mapB): unexpected report length (should be 1 line): %d",
			[]string{".fizz"}, len(report),
		)
	}

	report, err = Compare(nil, []string{".buzz"}, false, mapA, mapB)
	if err != nil {
		t.Errorf("Compare(nil, %q, false, mapA, mapB): unexpected error: %s", []string{".buzz"}, err)
	}
	if len(report) != 1 {
		t.Errorf(
			"Compare(nil, %q, false, mapA, mapB): unexpected report length (should be 1 line): %d",
			[]string{".buzz"}, len(report),
		)
	}

	report, err = Compare(nil, nil, true, mapA, mapC)
	if err != nil {
		t.Errorf("Compare(nil, nil, true, mapA, mapC): unexpected error: %s", err)
	}
	if len(report) != 1 {
		t.Errorf(
			"Compare(nil, nil, true, mapA, mapC): unexpected report length (should be 1 lines): %d",
			len(report),
		)
	}

	_, err = Compare(nil, nil, false, recMap, mapA)
	if err == nil {
		t.Errorf("Compare(nil, nil, false, recMap, mapA): expected error, got nil")
	}
}
