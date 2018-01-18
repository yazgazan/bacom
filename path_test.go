package backomp

import (
	"testing"
)

func TestMatchPath(t *testing.T) {
	for _, test := range []struct {
		Pattern  string
		Input    string
		Expected bool
	}{
		{"*", "foo", true},
		{"/*", "", true},
		{"/*", "/foo", true},
		{"/foo/*", "/foo/bar", true},
		{"/foo/bar", "/foo/bar", true},
		{"/foo", "/foo/bar", false},
		{"/foo/bar", "/foo", false},
		{"/**", "/foo/bar", true},
		{"/foo/**/bar", "/foo/bar", true},
		{"/foo/**/bar", "/foo/fizz/bar", true},
		{"/foo/**/bar", "/foo/fizz/buzz/bar", true},
		{"/foo*", "/foobar", true},
		{"/foo*", "/bar", false},
		{"/*foo", "/foo", true},
		{"/*foo", "/barfoo", true},
		{"/*foo", "/bar", false},
		{"", "", true},
	} {
		ok, err := MatchPath(test.Pattern, test.Input)
		if err != nil {
			t.Errorf("Match(%q, %q): unexpected error: %s", test.Pattern, test.Input, err)
		}
		if ok != test.Expected {
			t.Errorf("Match(%q, %q) = %v, expected %v", test.Pattern, test.Input, ok, test.Expected)
		}
	}
}
