# BaComp-tester

[![Go Report Card](https://goreportcard.com/badge/github.com/yazgazan/backomp)](https://goreportcard.com/report/github.com/yazgazan/backomp)
[![GoDoc](https://godoc.org/github.com/yazgazan/backomp?status.svg)](https://godoc.org/github.com/yazgazan/backomp)
[![Build Status](https://travis-ci.org/yazgazan/backomp.svg?branch=master)](https://travis-ci.org/yazgazan/backomp)
[![Coverage Status](https://coveralls.io/repos/github/yazgazan/backomp/badge.svg?branch=master)](https://coveralls.io/github/yazgazan/backomp?branch=master)

## Installing backomp

### Downloading the compiled binary

- Download the latest version of the binary: [releases](https://github.com/yazgazan/backomp/releases)
- extract the archive and place the `backomp` binary in your `$PATH`

### From source

- Have go 1.8 or greater installed: [golang.org](https://golang.org/doc/install)
- run `go get -u github.com/yazgazan/backomp/cmd/backomp`

## Usage

```
Usage: backomp [COMMAND] [OPTIONS]

COMMANDS:
    test    run existing tests
    import  import requests from HAR files
    list    lists tests information


Options for backomp test:
    -base-host string
        host for the base to compare to (leave empty to use saved tests versions)
    -base-use-https
        use https for requests to the base host
    -conf string
        configuration file (default "backcomp.json")
    -dir string
        directory containing the tests (default "backomp-tests")
    -q	Reduce standard output
    -save string
        save requests to target to the specified version
    -target-host string
        host for the target to compare (can include port) (default "localhost")
    -target-use-https
        use httpsfor the requests to the target host
    -v	print reasons
    -version value
        test version (default *)

Usage: backomp import  [SUB-COMMAND] [OPTIONS]

SUB-COMMANDS:
    har    import requests and responses from har files
    curl   save a request/response pair by providing curl-like arguments

Usage of backomp import har:
    -out string
        output directory (default ".")
    -v	verbose

Usage of backomp import curl:
    -H value
        Pass custom header to server (can be repeated)
    -X string
        Specify request command to use (default "GET")
    -compressed
        placeholder for curl's --compressed option
    -d value
        HTTP POST data
    -data value
        HTTP POST data
    -data-ascii value
        HTTP POST ASCII data
    -data-binary value
        HTTP POST binary data
    -data-raw value
        HTTP POST data, '@' allowed
    -dir string
        folder to save the request/response files in
    -name string
        name to save the request/response under (without the _req.txt suffix)
    -url string
        URL to work with
    -v	verbose

Usage of backomp list:
    -dir string
        folder containing the tests (default "backomp-tests")
    -l	prints detailed listing
    -version value
        constraint listing to these tests (default *)

```

## Examples

This command will run the tests located in the default `backomp-tests` folder, where the sub-directory matches the version constraint `<=1.x`.

```bash
./backomp test -target-host=localhost:1235 -version="<=1.x"
```

Alternatively, you can run the tests against a live/test endpoint instead of the saved responses:

```bash
./backomp test -target-host=localhost:1235 -base-host=example.org -version="<=1.x"
```

A configuration file can be used to specify path-based rules:

```bash
./backomp test -target-host=localhost:1235 -version="<=1.x" -conf=backomp-tests/ignore-bar.json
```
