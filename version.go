package backomp

import (
	"log"
	"os"
	"path"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

// Constraints is an interface for semver.Constraints
type Constraints interface {
	Validate(*semver.Version) (bool, []error)
}

// FindVersions returns the versions (folders) found that match the provided constraints
func FindVersions(dir string, verbose bool, constraints Constraints) (files []string, err error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "looking for versions in %q", dir)
	}
	defer handleClose(&err, f)

	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, errors.Wrapf(err, "looking for versions in %q", dir)
	}

	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		fname := fi.Name()

		v, err := parseVersion(verbose, fname)
		if err != nil {
			continue
		}

		valid, errs := constraints.Validate(v)
		if verbose {
			logErrors(fname, errs)
		}
		if !valid {
			continue
		}
		files = append(files, path.Join(dir, fname))
	}

	if len(files) == 0 {
		return nil, errors.Errorf("couldn't find versions matching %q in %s", constraints, dir)
	}

	return files, nil
}

func parseVersion(verbose bool, s string) (*semver.Version, error) {
	v, err := semver.NewVersion(s)
	if err != nil && verbose {
		log.Printf("invalid version name %q: %s", s, err)
	}

	return v, nil
}

func logErrors(fname string, errs []error) {
	log.Printf("validating version %q:", fname)
	for _, err := range errs {
		log.Println(err)
	}
}
