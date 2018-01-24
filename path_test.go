package bacom

import (
	"testing"
)

func TestMatchPath(t *testing.T) {
	for _, test := range []struct {
		Pattern  string
		Input    string
		Expected bool
		Error    bool
	}{
		{"*", "foo", true, false},
		{"/*", "", true, false},
		{"/*", "/foo", true, false},
		{"/foo/*", "/foo/bar", true, false},
		{"/foo/bar", "/foo/bar", true, false},
		{"/foo", "/foo/bar", false, false},
		{"/foo/bar", "/foo", false, false},
		{"/**", "/foo/bar", true, false},
		{"/foo/**/bar", "/foo/bar", true, false},
		{"/foo/**/bar", "/foo/fizz/bar", true, false},
		{"/foo/**/bar", "/foo/fizz/buzz/bar", true, false},
		{"/foo*", "/foobar", true, false},
		{"/foo*", "/bar", false, false},
		{"/*foo", "/foo", true, false},
		{"/*foo", "/barfoo", true, false},
		{"/*foo", "/bar", false, false},
		{"/foo/**/bar", "/bar/bar", false, false},
		{"/[-]/**/bar", "/bar/bar", false, true},
		{"[-]", "/bar", false, true},
		{"", "", true, false},
	} {
		ok, err := MatchPath(test.Pattern, test.Input)
		if !test.Error && err != nil {
			t.Errorf("Match(%q, %q): unexpected error: %s", test.Pattern, test.Input, err)
		}
		if test.Error && err == nil {
			t.Errorf("Match(%q, %q): expected error, got nil", test.Pattern, test.Input)
		}
		if ok != test.Expected {
			t.Errorf("Match(%q, %q) = %v, expected %v", test.Pattern, test.Input, ok, test.Expected)
		}
	}
}
