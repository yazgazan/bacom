package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/yazgazan/bacom"
	"gopkg.in/yaml.v2"
)

type pathConfFormat string

const (
	unknownFormat pathConfFormat = ""
	jsonPathConf  pathConfFormat = "json"
	yamlPathConf  pathConfFormat = "yaml"
	tomlPathConf  pathConfFormat = "toml"
)

type pathConf struct {
	Path     string
	Method   string
	Versions constraints
	JSON     jsonConf
	Headers  headersConf
}

type jsonConf struct {
	Ignore        []string
	IgnoreMissing []string `yaml:"ignore_missing"`
	IgnoreNull    bool     `yaml:"ignore_null"`
}

type headersConf struct {
	Ignore        []string
	IgnoreContent []string `yaml:"ignore_content"`
}

func getPathConf(verbose bool, conf []pathConf, version, method, path string) pathConf {
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
		ok, err = bacom.VersionMatch(verbose, c.Versions, version)
		if err != nil || !ok {
			continue
		}
		err = mergo.Merge(&pConf, c, mergo.WithOverride, mergo.WithAppendSlice)
		if err != nil {
			continue
		}
	}

	pConf.Path = ""
	pConf.Method = ""
	return pConf
}

func readPathConf(fname string, defaultConf []pathConf) ([]pathConf, error) {
	format := getPathConfFormat(fname)
	if format == unknownFormat {
		return nil, errors.Errorf(
			"invalid configuration format: %q. Supported formats are json, yaml and toml.",
			fname,
		)
	}
	readConf, err := pathConfReader(format)
	if err != nil {
		return nil, err
	}

	conf, err := readConf(fname)
	if err != nil {
		return conf, err
	}
	if conf == nil {
		return defaultConf, nil
	}

	return conf, nil
}

func pathConfReader(format pathConfFormat) (func(string) ([]pathConf, error), error) {
	switch format {
	default:
		return nil, errors.New("unknown configuration format")
	case jsonPathConf:
		return readJSONPathConf, nil
	case yamlPathConf:
		return readYAMLPathConf, nil
	case tomlPathConf:
		return readTOMLPathConf, nil
	}
}

func readJSONPathConf(fname string) (conf []pathConf, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, nil
	}
	defer handleClose(&err, f)

	err = json.NewDecoder(f).Decode(&conf)

	return conf, err
}

func readYAMLPathConf(fname string) (conf []pathConf, err error) {
	var yamlConf struct {
		Conf []pathConf
	}

	f, err := os.Open(fname)
	if err != nil {
		return nil, nil
	}
	defer handleClose(&err, f)

	err = yaml.NewDecoder(f).Decode(&yamlConf)

	return yamlConf.Conf, err
}

func readTOMLPathConf(fname string) (conf []pathConf, err error) {
	var tomlConf struct {
		Conf []pathConf
	}

	f, err := os.Open(fname)
	if err != nil {
		return nil, nil
	}
	defer handleClose(&err, f)

	_, err = toml.DecodeReader(f, &tomlConf)
	return tomlConf.Conf, err
}

func getPathConfFormat(fname string) pathConfFormat {
	switch strings.ToLower(filepath.Ext(fname)) {
	default:
		return unknownFormat
	case ".json":
		return jsonPathConf
	case ".yaml", ".yml":
		return yamlPathConf
	case ".toml":
		return tomlPathConf
	}
}
