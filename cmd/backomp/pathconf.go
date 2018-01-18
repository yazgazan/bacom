package main

import (
	"github.com/imdario/mergo"
	"github.com/yazgazan/backomp"
)

type pathConf struct {
	Path    string
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

func getPathConf(conf []pathConf, path string) pathConf {
	var pConf pathConf

	for _, c := range conf {
		ok, err := backomp.MatchPath(c.Path, path)
		if err != nil || !ok {
			continue
		}
		err = mergo.Merge(&pConf, c, mergo.WithOverride)
		if err != nil {
			continue
		}
	}

	pConf.Path = ""
	return pConf
}
