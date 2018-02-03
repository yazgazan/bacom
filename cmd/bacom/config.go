package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

const (
	defaultDir     = "bacom-tests"
	importCmdName  = "import"
	testCmdName    = "test"
	listCmdName    = "list"
	versionCmdName = "version"

	curlSubCmdName = "curl"
	harSubCmdName  = "har"
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
					"X-*", "Content-Length",
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
	Host       string
	UseHTTPS   bool
	PreProcess string
}

func printGlobalUsage() {
	bin := getBinaryName()
	fmt.Fprintf(
		os.Stderr,
		`Usage: %s [COMMAND] [OPTIONS]

COMMANDS:
    test    run existing tests
    import  import requests from HAR files
    list    lists tests information
    version print version information

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
	case testCmdName, importCmdName, listCmdName, versionCmdName:
		return strings.ToLower(cmd), args
	}

	return "", nil
}

func getBinaryName() string {
	if len(os.Args) == 0 {
		return "bacom"
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

	flags := flag.NewFlagSet(getBinaryName()+" "+testCmdName, flag.ExitOnError)

	flags.StringVar(&c.Dir, "dir", defaultDir, "directory containing the tests")
	flags.Var(&c.Constraints, "version", "test version")
	flags.StringVar(&c.Save, "save", "", "save requests to target to the specified version")
	flags.BoolVar(&c.Verbose, "v", false, "print reasons")
	flags.BoolVar(&c.Quiet, "q", false, "Reduce standard output")
	flags.StringVar(&c.PathsConfFile, "conf", "bacom.json", "configuration file")

	flags.StringVar(&c.Base.Host, "base-host", "", "host for the base to compare to (leave empty to use saved tests versions)")
	flags.BoolVar(&c.Base.UseHTTPS, "base-use-https", false, "use https for requests to the base host")
	flags.StringVar(&c.Base.PreProcess, "base-preprocess", "", "command used to pre-process requests sent to the base")
	flags.StringVar(&c.Target.Host, "target-host", "localhost", "host for the target to compare (can include port)")
	flags.BoolVar(&c.Target.UseHTTPS, "target-use-https", false, "use httpsfor the requests to the target host")
	flags.StringVar(&c.Target.PreProcess, "target-preprocess", "", "command used to pre-process requests sent to the target")
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
	Dir     string
	Files   []string
	Verbose bool
}

func parseImportHARFlags(args []string) (c importConf, err error) {
	flags := flag.NewFlagSet(getBinaryName()+" "+importCmdName+" "+harSubCmdName, flag.ExitOnError)

	flags.StringVar(&c.Dir, "out", ".", "output directory")
	flags.BoolVar(&c.Verbose, "v", false, "verbose")
	err = flags.Parse(args)
	if err != nil {
		return c, err
	}

	c.Files = flags.Args()

	return c, nil
}

type headers map[string][]string

func (h *headers) String() string {
	b := &bytes.Buffer{}

	err := http.Header(*h).Write(b)
	if err != nil {
		return ""
	}

	return b.String()
}

func (h *headers) Set(s string) error {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return errors.Errorf("invalid header %q", s)
	}
	if *h == nil {
		*h = make(map[string][]string)
	}

	http.Header(*h).Add(parts[0], parts[1])

	return nil
}

type dataFlag struct {
	io.ReadCloser
}

func (f dataFlag) String() string {
	return ""
}

func (f *dataFlag) Set(s string) error {
	if s == "" || s[0] != '@' {
		f.ReadCloser = ioutil.NopCloser(strings.NewReader(s))

		return nil
	}

	fname := s[1:]
	file, err := os.Open(fname)
	if err != nil {
		return errors.Wrapf(err, "opening %q", fname)
	}
	f.ReadCloser = file

	return nil
}

func (f dataFlag) Read(p []byte) (n int, err error) {
	if f.ReadCloser == nil {
		return 0, io.EOF
	}

	return f.ReadCloser.Read(p)
}

func (f dataFlag) Close() error {
	if f.ReadCloser == nil {
		return nil
	}

	return f.ReadCloser.Close()
}

type dataRawFlag struct {
	io.ReadCloser
}

func (f dataRawFlag) String() string {
	return ""
}

func (f *dataRawFlag) Set(s string) error {
	f.ReadCloser = ioutil.NopCloser(strings.NewReader(s))

	return nil
}

type curlConf struct {
	// CURL options
	Method  string
	URL     string
	Headers headers
	Data    dataFlag

	// bacom options
	Name    string
	Dir     string
	Verbose bool
}

func parseCurlFlags(args []string) (c curlConf, err error) {
	if len(args) != 0 && len(args[0]) != 0 && args[0][0] != '-' {
		args = append(args[1:], args[0])
	}

	flags := flag.NewFlagSet(getBinaryName()+" "+importCmdName+" "+curlSubCmdName, flag.ExitOnError)

	flags.StringVar(
		&c.Name, "name", "",
		"name to save the request/response under (without the _req.txt suffix)",
	)
	flags.StringVar(&c.Dir, "dir", "", "folder to save the request/response files in")
	flags.BoolVar(&c.Verbose, "v", false, "verbose")

	flags.StringVar(&c.Method, "X", "GET", "Specify request command to use")
	flags.StringVar(&c.URL, "url", "", "URL to work with")
	flags.Var(&c.Headers, "H", "Pass custom header to server (can be repeated)")

	flags.Var(&c.Data, "d", "HTTP POST data")
	flags.Var(&c.Data, "data", "HTTP POST data")
	flags.Var(&c.Data, "data-ascii", "HTTP POST ASCII data")
	flags.Var(&c.Data, "data-binary", "HTTP POST binary data")
	flags.Var((*dataRawFlag)(&c.Data), "data-raw", "HTTP POST data, '@' allowed")

	// Flags defined for compatibility purposes
	flags.Bool("compressed", false, "placeholder for curl's --compressed option")

	err = flags.Parse(args)
	if err != nil {
		return c, err
	}

	if len(flags.Args()) >= 2 {
		return c, errors.Errorf("expected one positional argument, got %d", len(flags.Args()))
	}
	if len(flags.Args()) == 1 {
		c.URL = flags.Args()[0]
	}

	return c, nil
}

type listConf struct {
	Dir         string
	Long        bool
	Constraints constraints
}

func parseListFlags(args []string) (c listConf, err error) {
	c = listConf{
		Constraints: defaultConstraints,
	}

	flags := flag.NewFlagSet(getBinaryName()+" "+listCmdName, flag.ExitOnError)

	flags.StringVar(&c.Dir, "dir", defaultDir, "folder containing the tests")
	flags.BoolVar(&c.Long, "l", false, "prints detailed listing")
	flags.Var(&c.Constraints, "version", "constraint listing to these tests")
	err = flags.Parse(args)

	return c, err
}
