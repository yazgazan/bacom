package backomp

import (
	"path"
)

// MatchPath uses path.Match to match a full path against a pattern.
// In addition to the path.Match pattern syntax, \** can be used to
// match any number of folder names.
func MatchPath(pattern, fpath string) (bool, error) {
	if pattern == "" || pattern[0] != '/' {
		pattern = "/" + pattern
	}
	if fpath == "" || fpath[0] != '/' {
		fpath = "/" + fpath
	}

	return matchPath(pattern, fpath)
}

func matchPath(pattern, fpath string) (bool, error) {
	pdir, pname := path.Split(path.Clean(pattern))
	fdir, fname := path.Split(path.Clean(fpath))

	if pname == "" || fname == "" {
		return true, nil
	}

	if pname == "**" {
		return matchPathUp(pdir, fpath)
	}

	if ok, err := path.Match(pname, fname); err != nil {
		return false, err
	} else if ok {
		return matchPath(pdir, fdir)
	}

	return false, nil
}

func matchPathUp(pattern, fpath string) (bool, error) {
	_, pname := path.Split(path.Clean(pattern))
	fdir, fname := path.Split(path.Clean(fpath))

	if pname == "" && fname == "" {
		return true, nil
	}
	if fname == "" {
		return false, nil
	}

	if ok, err := path.Match(pname, fname); err != nil {
		return false, err
	} else if ok {
		return matchPath(pattern, fpath)
	}

	return matchPathUp(pattern, fdir)
}
