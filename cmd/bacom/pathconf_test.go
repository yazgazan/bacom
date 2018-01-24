package main

import (
	"reflect"
	"testing"
)

func TestGetPathConf(t *testing.T) {
	for _, test := range []struct {
		path     string
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
	} {
		pConf := getPathConf(test.conf, test.path)

		if !reflect.DeepEqual(pConf, test.expected) {
			t.Errorf("getPathConf(%+v, %q) = %+v, expected %+v", test.conf, test.path, pConf, test.expected)
		}
	}
}
