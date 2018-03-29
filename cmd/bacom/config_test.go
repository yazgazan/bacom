package main

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

func TestConstraints(t *testing.T) {
	var c constraints

	v := "v0.1.x"
	err := c.Set(v)
	if err != nil {
		t.Errorf("constraints.Set(%q): unexpected error: %s", v, err)
	}
	if c.String() != v {
		t.Errorf("constraints.Set(%q).String() = %q, expected %q", v, c.String(), v)
	}
}

func TestStringsFlag(t *testing.T) {
	var flags stringsFlag

	s := "foo,bar"
	expected := stringsFlag{"foo", "bar"}
	err := flags.Set(s)
	if err != nil {
		t.Errorf("stringsFlag.Set(%q): unexpected error %s", s, err)
	}
	if !reflect.DeepEqual(flags, expected) {
		t.Errorf("stringsFlag.Set(%q) = %q, expected %q", s, flags, expected)
	}
	if flags.String() != s {
		t.Errorf("stringsFlag.Set(%q).String() = %q, expected %q", s, flags.String(), s)
	}

	s = "fizz,buzz,"
	expectedString := "foo,bar,fizz,buzz"
	expected = append(expected, "fizz", "buzz")
	err = flags.Set(s)
	if err != nil {
		t.Errorf("%q.Set(%q): unexpected error %s", flags, s, err)
	}
	if !reflect.DeepEqual(flags, expected) {
		t.Errorf("%q.Set(%q) = %q, expected %q", flags, s, flags, expected)
	}
	if flags.String() != expectedString {
		t.Errorf("%q.Set(%q).String() = %q, expected %q", flags, s, flags.String(), expectedString)
	}
}

func TestRegexesFlag(t *testing.T) {
	var flags regexesFlag

	s := ".+\\.foo\\.org"
	expected := regexesFlag{
		regexp.MustCompile(s),
	}
	err := flags.Set(s)
	if err != nil {
		t.Errorf("regexesFlag.Set(%q): unexpected error: %s", s, err)
	}
	if !reflect.DeepEqual(flags, expected) {
		t.Errorf("regexesFlag.Set(%q) = %q, expected %q", s, flags, expected)
	}
	expectedString := fmt.Sprintf("%q", []string{s})
	if flags.String() != expectedString {
		t.Errorf("regexesFlag.Set(%q).String() = %q, expected %q", s, flags.String(), expectedString)
	}
}
