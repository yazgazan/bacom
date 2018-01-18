package backomp

import "testing"

func TestIsRequestFilename(t *testing.T) {
	for _, test := range []struct {
		In    string
		Value bool
	}{
		{"_req.txt", true},
		{"", false},
		{"foo_req.txt", true},
		{"foo_req1.txt", true},
		{"foo_req2.txt", true},
		{"foo_req42.txt", true},
		{"bar_REQ.txt", false},
		{"foo.txt", false},
		{"_req.txt.txt", false},
		{"_req.txt_req.txt", true},
	} {
		v := isRequestFilename(test.In)
		if v != test.Value {
			t.Errorf("isRequestFilename(%q) = %v, expected %v", test.In, v, test.Value)
		}
	}
}

func TestGetResponseFilename(t *testing.T) {
	for _, test := range []struct {
		In       string
		Expected string
		Err      error
	}{
		{"_req.txt", "_resp.txt", nil},
		{"foo_req.txt", "foo_resp.txt", nil},
		{"foo_req1.txt", "foo_resp1.txt", nil},
		{"foo_req2.txt", "foo_resp2.txt", nil},
		{"foo_req24.txt", "foo_resp24.txt", nil},
		{"_req.txt_req.txt", "_req.txt_resp.txt", nil},
		{"foo.txt", "", ErrReqInvalidName},
		{"_req.txt.txt", "", ErrReqInvalidName},
		{"", "", ErrReqInvalidName},
	} {
		v, err := GetResponseFilename(test.In)
		if v != test.Expected {
			t.Errorf("getResponseFilename(%q) = %q, expected %q", test.In, v, test.Expected)
		}
		if err == nil && test.Err != nil {
			t.Errorf("getResponseFilename(%q): got nil, expected error %q", test.In, test.Err)
		}
		if err != nil && test.Err == nil {
			t.Errorf("getResponseFilename(%q): unexpected error %q", test.In, err)
		}
	}
}
