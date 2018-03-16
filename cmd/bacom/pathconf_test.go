package main

import (
	"reflect"
	"testing"
)

func TestGetPathConf(t *testing.T) {
	for _, test := range []struct {
		path     string
		method   string
		conf     []pathConf
		expected pathConf
	}{
		{
			path:     "/foo",
			conf:     []pathConf{},
			expected: pathConf{},
		},
		{
			path: "/foo",
			conf: []pathConf{
				{
					Path: "**",
					JSON: jsonConf{
						Ignore:     []string{".error"},
						IgnoreNull: true,
					},
					Headers: headersConf{
						Ignore: []string{"Content-Type"},
					},
				},
			},
			expected: pathConf{
				JSON: jsonConf{
					Ignore:     []string{".error"},
					IgnoreNull: true,
				},
				Headers: headersConf{
					Ignore: []string{"Content-Type"},
				},
			},
		},
		{
			path: "/foo",
			conf: []pathConf{
				{
					Path: "**",
					JSON: jsonConf{
						Ignore:     []string{".error"},
						IgnoreNull: true,
					},
					Headers: headersConf{
						Ignore: []string{"Content-Type"},
					},
				},
				{
					Path: "/foo",
					JSON: jsonConf{
						Ignore:     []string{".reason"},
						IgnoreNull: false,
					},
					Headers: headersConf{
						Ignore: []string{"Accept-Language"},
					},
				},
				{
					Path: "/bar",
					JSON: jsonConf{
						Ignore:     []string{".result"},
						IgnoreNull: true,
					},
					Headers: headersConf{
						Ignore: []string{"Pragma"},
					},
				},
			},
			expected: pathConf{
				JSON: jsonConf{
					Ignore:     []string{".error", ".reason"},
					IgnoreNull: true,
				},
				Headers: headersConf{
					Ignore: []string{"Content-Type", "Accept-Language"},
				},
			},
		},
		{
			path:   "/foo",
			method: "POST",
			conf: []pathConf{
				{
					Path: "**",
					JSON: jsonConf{
						Ignore:     []string{".error"},
						IgnoreNull: true,
					},
					Headers: headersConf{
						Ignore: []string{"Content-Type"},
					},
				},
				{
					Path:   "/foo",
					Method: "GET",
					JSON: jsonConf{
						Ignore:     []string{".reason"},
						IgnoreNull: false,
					},
					Headers: headersConf{
						Ignore: []string{"Accept-Language"},
					},
				},
				{
					Path: "/bar",
					JSON: jsonConf{
						Ignore:     []string{".result"},
						IgnoreNull: true,
					},
					Headers: headersConf{
						Ignore: []string{"Pragma"},
					},
				},
			},
			expected: pathConf{
				JSON: jsonConf{
					Ignore:     []string{".error"},
					IgnoreNull: true,
				},
				Headers: headersConf{
					Ignore: []string{"Content-Type"},
				},
			},
		},
		{
			path:   "/foo",
			method: "GET",
			conf: []pathConf{
				{
					Path: "**",
					JSON: jsonConf{
						Ignore:     []string{".error"},
						IgnoreNull: true,
					},
					Headers: headersConf{
						Ignore: []string{"Content-Type"},
					},
				},
				{
					Path:   "/foo",
					Method: "GET",
					JSON: jsonConf{
						Ignore:     []string{".reason"},
						IgnoreNull: false,
					},
					Headers: headersConf{
						Ignore: []string{"Accept-Language"},
					},
				},
				{
					Path: "/bar",
					JSON: jsonConf{
						Ignore:     []string{".result"},
						IgnoreNull: true,
					},
					Headers: headersConf{
						Ignore: []string{"Pragma"},
					},
				},
			},
			expected: pathConf{
				JSON: jsonConf{
					Ignore:     []string{".error", ".reason"},
					IgnoreNull: true,
				},
				Headers: headersConf{
					Ignore: []string{"Content-Type", "Accept-Language"},
				},
			},
		},
	} {
		pConf := getPathConf(test.conf, test.method, test.path)

		if !reflect.DeepEqual(pConf, test.expected) {
			t.Errorf("getPathConf(%+v, %q, %q) = %+v, expected %+v", test.conf, test.method, test.path, pConf, test.expected)
		}
	}
}

func TestGetPathConfFormat(t *testing.T) {
	for _, test := range []struct {
		fname    string
		expected pathConfFormat
	}{
		{"/foo/bar.json", jsonPathConf},
		{"../foo.json", jsonPathConf},
		{"foo.json", jsonPathConf},
		{"foo.yaml", yamlPathConf},
		{"foo.yml", yamlPathConf},
		{"foo.toml", tomlPathConf},
		{"foo.txt", unknownFormat},
		{"foo", unknownFormat},
		{"", unknownFormat},
	} {
		got := getPathConfFormat(test.fname)
		if got != test.expected {
			t.Errorf("getPathConfFormat(%q) = %s, expected %s", test.fname, got, test.expected)
		}
	}
}

var expectedTestConf = []pathConf{
	{
		Path: "**",
		Headers: headersConf{
			Ignore: []string{"Connection"},
			IgnoreContent: []string{
				"Age", "Content-MD5", "Content-Range", "Date",
				"Expires", "Last-Modified", "Public-Key-Pins",
				"Server", "Set-Cookie", "Etag", "Retry-After",
				"X-*", "Content-Length",
			},
		},
	},
	{
		Path:   "/api",
		Method: "GET",
		JSON: jsonConf{
			Ignore:        []string{".Results[].Bar"},
			IgnoreMissing: []string{".Results[].Foo"},
		},
		Headers: headersConf{
			IgnoreContent: []string{"Cache-Control"},
		},
	},
}

func TestReadPathConf(t *testing.T) {
	defaultConf := []pathConf{
		{
			Path:   "**",
			Method: "POST",
			JSON: jsonConf{
				Ignore: []string{".Foo"},
			},
		},
	}
	fname := "/tmp/doesnotexist.json"
	conf, err := readPathConf(fname, defaultConf)
	if err != nil {
		t.Errorf("readPathConf(%q, defaultConf): unexpected error: %s", fname, err)
	}
	if !reflect.DeepEqual(conf, defaultConf) {
		t.Errorf("readPathConf(%q, defaultConf) = %+v, expected %+v", fname, conf, defaultConf)
	}

	fname = "/tmp/unkownformat.txt"
	_, err = readPathConf(fname, defaultConf)
	if err == nil {
		t.Errorf("readPathConf(%q, defaultConf): expected error, got nil", fname)
	}

	fname = "../../bacom-tests/ignore-bar.json"
	conf, err = readPathConf(fname, defaultConf)
	if err != nil {
		t.Errorf("readPathConf(%q, defaultConf): unexpected error: %s", fname, err)
	}
	if !reflect.DeepEqual(conf, expectedTestConf) {
		t.Errorf("readPathConf(%q, defaultConf) = %+v, expected %+v", fname, conf, expectedTestConf)
	}
}

func TestReadJSONPathConf(t *testing.T) {
	fname := "/tmp/doesnotexist.json"
	_, err := readJSONPathConf(fname)
	if err != nil {
		t.Errorf("readJSONPathConf(%q): unexpected error: %s", fname, err)
	}

	fname = "../../bacom-tests/ignore-bar.json"
	conf, err := readJSONPathConf(fname)
	if err != nil {
		t.Errorf("readJSONPathConf(%q): unexpected error: %s", fname, err)
	}
	if !reflect.DeepEqual(conf, expectedTestConf) {
		t.Errorf("readJSONPathConf(%q) = %+v, expected %+v", fname, conf, expectedTestConf)
	}
}

func TestReadYAMLPathConf(t *testing.T) {
	fname := "/tmp/doesnotexist.yaml"
	_, err := readYAMLPathConf(fname)
	if err != nil {
		t.Errorf("readYAMLPathConf(%q): unexpected error: %s", fname, err)
	}

	fname = "../../bacom-tests/ignore-bar.yaml"
	conf, err := readYAMLPathConf(fname)
	if err != nil {
		t.Errorf("readYAMLPathConf(%q): unexpected error: %s", fname, err)
	}
	if !reflect.DeepEqual(conf, expectedTestConf) {
		t.Errorf("readYAMLPathConf(%q) = %+v, expected %+v", fname, conf, expectedTestConf)
	}
}

func TestReadTOMLPathConf(t *testing.T) {
	fname := "/tmp/doesnotexist.toml"
	_, err := readTOMLPathConf(fname)
	if err != nil {
		t.Errorf("readTOMLPathConf(%q): unexpected error: %s", fname, err)
	}

	fname = "../../bacom-tests/ignore-bar.toml"
	conf, err := readTOMLPathConf(fname)
	if err != nil {
		t.Errorf("readTOMLPathConf(%q): unexpected error: %s", fname, err)
	}
	if !reflect.DeepEqual(conf, expectedTestConf) {
		t.Errorf("readTOMLPathConf(%q) = %+v, expected %+v", fname, conf, expectedTestConf)
	}
}

func TestPathConfReader(t *testing.T) {
	for _, test := range []struct {
		format      pathConfFormat
		expected    func(string) ([]pathConf, error)
		expectError bool
	}{
		{jsonPathConf, readJSONPathConf, false},
		{yamlPathConf, readYAMLPathConf, false},
		{tomlPathConf, readTOMLPathConf, false},
		{unknownFormat, nil, true},
	} {
		fn, err := pathConfReader(test.format)
		if err != nil && !test.expectError {
			t.Errorf("pathConfReader(%s): unexpected error: %s", test.format, err)
			continue
		}
		if err == nil && test.expectError {
			t.Errorf("pathConfReader(%s): expected error, got nil", test.format)
			continue
		}
		if !test.expectError && fn == nil {
			t.Errorf("pathConfReader(%s): expected reader, got nil", test.format)
		}
	}
}
