package main

import (
	"strings"

	"github.com/imdario/mergo"
	"github.com/yazgazan/bacom"
)

type pathConf struct {
	Path    string
	Method  string
	JSON    jsonConf
	Headers headersConf
}

type jsonConf struct {
	Ignore        []string
	IgnoreMissing []string
	IgnoreNull    bool
}

type headersConf struct {
	Ignore        []string
	IgnoreContent []string
}

func getPathConf(conf []pathConf, method, path string) pathConf {
	var pConf pathConf

	method = strings.ToLower(method)
	for _, c := range conf {
		ok, err := bacom.MatchPath(c.Path, path)
		if err != nil || !ok {
			continue
		}
		if c.Method != "" && strings.ToLower(c.Method) != method {
			continue
		}
		err = mergo.Merge(&pConf, c, mergo.WithOverride)
		if err != nil {
			continue
		}
	}

	pConf.Path = ""
	pConf.Method = ""
	return pConf
}
