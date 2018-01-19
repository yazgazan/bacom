package backomp

import (
	"net/http"
	"testing"
)

func TestCompareHeaders(t *testing.T) {
	headerA := http.Header{
		"Content-Type": {"application/json"},
		"Date":         {"some date"},
	}
	headerB := http.Header{
		"Content-Type": {"application/json"},
	}
	headerC := http.Header{
		"Content-Type": {"application/json"},
		"Date":         {"some other date"},
	}

	report, err := CompareHeaders(nil, nil, headerA, headerA)
	if err != nil {
		t.Errorf("CompareHeaders(nil, nil, headerA, headerA): unexpected error: %s", err)
	}
	if len(report) != 0 {
		t.Errorf("CompareHeaders(nil, nil, headerA, headerA): invalid report. Expected 0 lines, got %d", len(report))
	}

	report, err = CompareHeaders(nil, nil, headerA, headerB)
	if err != nil {
		t.Errorf("CompareHeaders(nil, nil, headerA, headerB): unexpected error: %s", err)
	}
	if len(report) != 1 {
		t.Errorf("CompareHeaders(nil, nil, headerA, headerB): invalid report. Expected 1 line, got %d", len(report))
	}

	report, err = CompareHeaders([]string{"Date"}, nil, headerA, headerB)
	if err != nil {
		t.Errorf("CompareHeaders(%q, nil, headerA, headerB): unexpected error: %s", []string{"Date"}, err)
	}
	if len(report) != 0 {
		t.Errorf("CompareHeaders(%q, nil, headerA, headerB): invalid report. Expected 0 lines, got %d", []string{"Date"}, len(report))
	}

	report, err = CompareHeaders(nil, []string{"Date"}, headerA, headerC)
	if err != nil {
		t.Errorf("CompareHeaders(nil, %q, headerA, headerC): unexpected error: %s", []string{"Date"}, err)
	}
	if len(report) != 0 {
		t.Errorf("CompareHeaders(nil, %q, headerA, headerC): invalid report. Expected 0 lines, got %d", []string{"Date"}, len(report))
	}

	report, err = CompareHeaders(nil, nil, headerA, headerC)
	if err != nil {
		t.Errorf("CompareHeaders(nil, nil, headerA, headerC): unexpected error: %s", err)
	}
	if len(report) != 2 {
		t.Errorf("CompareHeaders(nil, nil, headerA, headerC): invalid report. Expected 2 lines, got %d", len(report))
	}

	_, err = CompareHeaders(nil, []string{"[-]"}, headerA, headerC)
	if err == nil {
		t.Errorf("CompareHeaders(nil, %q, headerA, headerC): expected error, got nil", []string{"[-]"})
	}

	_, err = CompareHeaders([]string{"[-]"}, nil, headerA, headerC)
	if err == nil {
		t.Errorf("CompareHeaders(%q, nil, headerA, headerC): expected error, got nil", []string{"[-]"})
	}
}
