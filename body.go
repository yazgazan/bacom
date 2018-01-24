package bacom

import (
	"github.com/yazgazan/jaydiff/diff"
)

// Compare returns a list of differences between two json objects.
// The ignore and ignoreMissing parameters are a list of JSON paths that should be ignored.
// If ignoreNull is true, nil values in the lhs won't be tested.
func Compare(ignore, ignoreMissing []string, ignoreNull bool, lhs, rhs interface{}) ([]string, error) {
	d, err := diff.Diff(lhs, rhs)
	if err != nil {
		return nil, err
	}

	d = IgnorePrunner(ignore).Prune(d)
	d = IgnoreMissingPrunner(ignoreMissing).Prune(d)
	d = Prune(d, ignoreNull)

	return diff.Report(d, diff.Output{
		Indent:    "\t",
		ShowTypes: false,
		Colorized: true,
	})
}
