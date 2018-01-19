package backomp

import (
	"net/http"
	"path"

	"github.com/fatih/color"
)

// CompareHeaders returns a list of differences between two http.Header.
// ignore and ignoreContent are expected to be normalized http headers names.
func CompareHeaders(ignore, ignoreContent []string, lhs, rhs http.Header) ([]string, error) {
	var results []string

	for k := range lhs {
		if ok, err := containsPattern(ignore, k); err != nil {
			return results, err
		} else if ok {
			continue
		}
		if _, ok := rhs[k]; !ok {
			results = append(results, missingHeader(k, lhs.Get(k)))
			continue
		}
		if ok, err := containsPattern(ignoreContent, k); err != nil {
			return results, err
		} else if ok {
			continue
		}
		if lhs.Get(k) != rhs.Get(k) {
			results = append(results,
				missingHeader(k, lhs.Get(k)),
				excessHeader(k, rhs.Get(k)),
			)
		}
	}

	return results, nil
}

func missingHeader(name, value string) string {
	return "- (Header) " + name + ": " + color.RedString("%s", value)
}

func excessHeader(name, value string) string {
	return "+ (Header) " + name + ": " + color.GreenString("%s", value)
}

func containsPattern(patterns []string, needle string) (bool, error) {
	for _, pattern := range patterns {
		match, err := path.Match(pattern, needle)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}

	return false, nil
}
