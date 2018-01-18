package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

const (
	defaultDir = "backomp-tests"
)

var (
	defaultConstraints = newConstraintMustParse("*")
	defaultPathsConfig = []pathConf{
		pathConf{
			Path: "**",
			Headers: headersConf{
				Ignore: []string{
					"Set-Cookie",
				},
			},
		},
	}
)

func newConstraintMustParse(s string) constraints {
	c, err := semver.NewConstraint(s)
	if err != nil {
		panic(err)
	}

	return constraints{
		Constraints: c,
		str:         "*",
	}
}

type conf struct {
	Dir           string
	Constraints   constraints
	Save          string
	Verbose       bool
	Quiet         bool
	PathsConfFile string

	Base   targetConf
	Target targetConf
	Paths  []pathConf
}

type constraints struct {
	*semver.Constraints
	str string
}

func (c *constraints) Set(s string) error {
	var err error

	c.str = s
	c.Constraints, err = semver.NewConstraint(s)

	return err
}

func (c constraints) String() string {
	return c.str
}

type targetConf struct {
	Host     string
	UseHTTPS bool
}

func parseFlags() (c conf, err error) {
	c = conf{
		Constraints: defaultConstraints,
		Paths:       defaultPathsConfig,
	}

	flag.StringVar(&c.Dir, "dir", defaultDir, "directory containing the tests")
	flag.Var(&c.Constraints, "version", "test version")
	flag.StringVar(&c.Save, "save", "", "save requests to target to the specified version")
	flag.BoolVar(&c.Verbose, "v", false, "print reasons")
	flag.BoolVar(&c.Quiet, "q", false, "Reduce standard output")
	flag.StringVar(&c.PathsConfFile, "conf", "backcomp.json", "configuration file")

	flag.StringVar(&c.Base.Host, "base-host", "", "host for the base to compare to (leave empty to use saved tests versions)")
	flag.BoolVar(&c.Base.UseHTTPS, "base-use-https", false, "use https for requests to the base host")
	flag.StringVar(&c.Target.Host, "target-host", "localhost", "host for the target to compare (can include port)")
	flag.BoolVar(&c.Target.UseHTTPS, "target-use-https", false, "use httpsfor the requests to the target host")
	flag.Parse()

	if c.Verbose && c.Quiet {
		return c, errors.New("conflicting -v and -q")
	}
	if c.PathsConfFile == "" {
		return c, nil
	}

	f, err := os.Open(c.PathsConfFile)
	if err != nil {
		return c, nil
	}
	defer handleClose(&err, f)

	err = json.NewDecoder(f).Decode(&c.Paths)

	return c, errors.Wrapf(err, "parsing configuration file %q", c.PathsConfFile)
}
