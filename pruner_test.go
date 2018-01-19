package backomp

import (
	"testing"

	"github.com/yazgazan/jaydiff/diff"
)

func TestPrune(t *testing.T) {
	mapA := map[string]interface{}{
		"foo":  2,
		"bar":  42,
		"fizz": []interface{}{1, 2, 3},
	}
	mapB := map[string]interface{}{
		"foo":  3,
		"bar":  23,
		"fizz": []interface{}{2, 2, 5, 6},
		"buzz": 55,
	}
	mapC := map[string]interface{}{
		"foo":  3,
		"bar":  nil,
		"fizz": []interface{}{2, 2, 5, 6},
	}

	d, err := diff.Diff(mapA, mapB)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapB): unexpected error: %s", err)
		return
	}
	d = Prune(d, false)
	if d.Diff() != diff.Identical {
		t.Errorf(
			"Prune(diff.Diff(mapA, mapB), false).Diff() = %s, expected %s",
			d.Diff(), diff.Identical,
		)
	}

	d, err = diff.Diff(mapC, mapA)
	if err != nil {
		t.Errorf("diff.Diff(mapC, mapA): unexpected error: %s", err)
		return
	}
	d = Prune(d, true)
	if d.Diff() != diff.Identical {
		t.Errorf(
			"Prune(diff.Diff(mapC, mapA), true).Diff() = %s, expected %s",
			d.Diff(), diff.Identical,
		)
	}

	d, err = diff.Diff(mapA, mapC)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapC): unexpected error: %s", err)
		return
	}
	d = Prune(d, true)
	if d.Diff() != diff.Identical {
		t.Errorf(
			"Prune(diff.Diff(mapA, mapC), true).Diff() = %s, expected %s",
			d.Diff(), diff.Identical,
		)
	}

	d, err = diff.Diff(mapA, mapC)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapC): unexpected error: %s", err)
		return
	}
	d = Prune(d, false)
	if d.Diff() != diff.ContentDiffer {
		t.Errorf(
			"Prune(diff.Diff(mapA, mapC), false).Diff() = %s, expected %s",
			d.Diff(), diff.ContentDiffer,
		)
	}
}

func TestIgnorePrunner(t *testing.T) {
	mapA := map[string]interface{}{
		"foo":  42,
		"fizz": 23,
	}
	mapB := map[string]interface{}{
		"foo":  "bar",
		"fizz": 23,
	}

	d, err := diff.Diff(mapA, mapB)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapB): unexpected error: %s", err)
		return
	}
	prunner := IgnorePrunner{".foo"}
	d = prunner.Prune(d)
	if d.Diff() != diff.Identical {
		t.Errorf(
			"IgnorePrunner{%q}.Prune(diff.Diff(mapA, mapB)).Diff() = %s, expected %s",
			prunner, d.Diff(), diff.Identical,
		)
	}

	d, err = diff.Diff(mapA, mapB)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapB): unexpected error: %s", err)
		return
	}
	prunner = IgnorePrunner{".fizz"}
	d = prunner.Prune(d)
	if d.Diff() != diff.ContentDiffer {
		t.Errorf(
			"IgnorePrunner{%q}.Prune(diff.Diff(mapA, mapB)).Diff() = %s, expected %s",
			prunner, d.Diff(), diff.ContentDiffer,
		)
	}

	d, err = diff.Diff(mapA, mapB)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapB): unexpected error: %s", err)
		return
	}
	prunner = IgnorePrunner{}
	d = prunner.Prune(d)
	if d.Diff() != diff.ContentDiffer {
		t.Errorf(
			"IgnorePrunner{%q}.Prune(diff.Diff(mapA, mapB)).Diff() = %s, expected %s",
			prunner, d.Diff(), diff.ContentDiffer,
		)
	}
}

func TestIgnoreMissingPrunner(t *testing.T) {
	mapA := map[string]interface{}{
		"foo": 42,
		"bar": 23,
	}
	mapB := map[string]interface{}{
		"foo": 42,
	}

	d, err := diff.Diff(mapA, mapB)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapB): unexpected error: %s", err)
		return
	}
	prunner := IgnoreMissingPrunner{".bar"}
	d = prunner.Prune(d)
	if d.Diff() != diff.Identical {
		t.Errorf(
			"IgnoreMissingPrunner{%q}.Prune(diff.Diff(mapA, mapB)).Diff() = %s, expected %s",
			prunner, d.Diff(), diff.Identical,
		)
	}

	d, err = diff.Diff(mapA, mapB)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapB): unexpected error: %s", err)
		return
	}
	prunner = IgnoreMissingPrunner{".foo"}
	d = prunner.Prune(d)
	if d.Diff() != diff.ContentDiffer {
		t.Errorf(
			"IgnoreMissingPrunner{%q}.Prune(diff.Diff(mapA, mapB)).Diff() = %s, expected %s",
			prunner, d.Diff(), diff.ContentDiffer,
		)
	}

	d, err = diff.Diff(mapA, mapB)
	if err != nil {
		t.Errorf("diff.Diff(mapA, mapB): unexpected error: %s", err)
		return
	}
	prunner = IgnoreMissingPrunner{}
	d = prunner.Prune(d)
	if d.Diff() != diff.ContentDiffer {
		t.Errorf(
			"IgnoreMissingPrunner{%q}.Prune(diff.Diff(mapA, mapB)).Diff() = %s, expected %s",
			prunner, d.Diff(), diff.ContentDiffer,
		)
	}
}
