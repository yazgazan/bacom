package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

const (
	defaultDir = "backomp-tests"
)

var (
	defaultConstraints = newConstraintMustParse("*")
	defaultPathsConfig = []pathConf{
		{
			Path: "**",
			Headers: headersConf{
				Ignore: []string{
					"Connection",
				},
				IgnoreContent: []string{
					"Age", "Content-MD5", "Content-Range", "Date",
					"Expires", "Last-Modified", "Public-Key-Pins",
					"Server", "Set-Cookie", "Etag", "Retry-After",
					"X-*",
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

func printGlobalUsage() {
	bin := getBinaryName()
	fmt.Fprintf(
		os.Stderr,
		`Usage: %s [COMMAND] [OPTIONS]

COMMANDS:
    test    run existing tests
    import  import requests from HAR files

Note:
    "%s COMMAND -h" to get an overview of each command's flags

`,
		bin, bin,
	)

}

func getCommand() (cmd string, args []string) {
	args = os.Args[1:]
	if len(args) == 0 {
		printGlobalUsage()
		os.Exit(2)
	}
	cmd = args[0]
	args = args[1:]

	switch strings.ToLower(cmd) {
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", cmd)
		os.Exit(2)
	case "test", "import":
		return strings.ToLower(cmd), args
	}

	return "", nil
}

func getBinaryName() string {
	if len(os.Args) == 0 {
		return "backomp"
	}

	return os.Args[0]
}

type testConf struct {
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

func parseTestFlags(args []string) (c testConf, err error) {
	c = testConf{
		Constraints: defaultConstraints,
		Paths:       defaultPathsConfig,
	}

	flags := flag.NewFlagSet(getBinaryName()+" test", flag.ExitOnError)

	flags.StringVar(&c.Dir, "dir", defaultDir, "directory containing the tests")
	flags.Var(&c.Constraints, "version", "test version")
	flags.StringVar(&c.Save, "save", "", "save requests to target to the specified version")
	flags.BoolVar(&c.Verbose, "v", false, "print reasons")
	flags.BoolVar(&c.Quiet, "q", false, "Reduce standard output")
	flags.StringVar(&c.PathsConfFile, "conf", "backcomp.json", "configuration file")

	flags.StringVar(&c.Base.Host, "base-host", "", "host for the base to compare to (leave empty to use saved tests versions)")
	flags.BoolVar(&c.Base.UseHTTPS, "base-use-https", false, "use https for requests to the base host")
	flags.StringVar(&c.Target.Host, "target-host", "localhost", "host for the target to compare (can include port)")
	flags.BoolVar(&c.Target.UseHTTPS, "target-use-https", false, "use httpsfor the requests to the target host")
	err = flags.Parse(args)
	if err != nil {
		return c, err
	}

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

type importConf struct {
	Dir   string
	Files []string
}

func parseImportFlags(args []string) (c importConf, err error) {
	c = importConf{}

	flags := flag.NewFlagSet(getBinaryName()+" import", flag.ExitOnError)

	flags.StringVar(&c.Dir, "out", ".", "output directory")
	err = flags.Parse(args)
	if err != nil {
		return c, err
	}

	c.Files = flags.Args()

	return c, nil
}
