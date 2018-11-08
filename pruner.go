package bacom

import (
	"github.com/pkg/errors"
	"github.com/yazgazan/jaydiff/diff"
	"github.com/yazgazan/jaydiff/jpath"
)

// Prune returns a diff.Differ, stripping the diff tree of the following differences:
//
// - Excess keys in the right hand side
// - Excess and Missing values in slices
// - Values of the same type with different content (excluding slices and maps)
// - Type difference where null is involved
func Prune(d diff.Differ, ignoreNull bool) diff.Differ {
	d, err := diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		switch {
		case diff.IsScalar(d) && d.Diff() == diff.ContentDiffer:
			fallthrough
		case ignoreNull && isNil(d):
			fallthrough
		case diff.IsExcess(d):
			fallthrough
		case diff.IsSlice(parent) && diff.IsMissing(d):
			return diff.Ignore()
		}

		return nil, nil
	})
	if err != nil {
		panic(errors.Wrap(err, "impossible error while pruning"))
	}

	return d
}

func isNil(d diff.Differ) bool {
	lhs, err := diff.LHS(d)
	if err != nil {
		return false
	}
	if lhs == nil {
		return true
	}

	rhs, err := diff.RHS(d)
	if err != nil {
		return false
	}
	if rhs == nil {
		return true
	}

	return false
}

// IgnorePrunner can be used to ignore json paths in a diff tree
type IgnorePrunner []string

// Prune Removes ignored diff branches from the diff tree
func (p IgnorePrunner) Prune(d diff.Differ) diff.Differ {
	if len(p) == 0 {
		return d
	}
	d, err := diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		if pathMatches(p, path) {
			return diff.Ignore()
		}
		return nil, nil
	})
	if err != nil {
		panic(errors.Wrap(err, "impossible error while pruning"))
	}

	return d
}

// IgnoreMissingPrunner can be used to ignore missing json paths (from the right hand side) in a diff tree
type IgnoreMissingPrunner []string

// Prune Removes ignored diff branches from the diff tree
func (p IgnoreMissingPrunner) Prune(d diff.Differ) diff.Differ {
	if len(p) == 0 {
		return d
	}

	d, err := diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		if !diff.IsMissing(d) {
			return nil, nil
		}

		if pathMatches(p, path) {
			return diff.Ignore()
		}

		return nil, nil
	})
	if err != nil {
		panic(errors.Wrap(err, "impossible error while pruning"))
	}

	return d
}

func pathMatches(paths []string, path string) bool {
	for _, p := range paths {
		if jpath.HasSuffix(path, p) {
			return true
		}
	}

	return false
}
