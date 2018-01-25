# Bacom

Pronounced like bacon, but for compatibility.

Bacom will help you test JSON apis for backward-compatibility breaking changes.

[![Go Report Card](https://goreportcard.com/badge/github.com/yazgazan/bacom)](https://goreportcard.com/report/github.com/yazgazan/bacom)
[![GoDoc](https://godoc.org/github.com/yazgazan/bacom?status.svg)](https://godoc.org/github.com/yazgazan/bacom)
[![Build Status](https://travis-ci.org/yazgazan/bacom.svg?branch=master)](https://travis-ci.org/yazgazan/bacom)
[![Coverage Status](https://coveralls.io/repos/github/yazgazan/bacom/badge.svg?branch=master)](https://coveralls.io/github/yazgazan/bacom?branch=master)

## Installing bacom

### Downloading the compiled binary

- Download the latest version of the binary: [releases](https://github.com/yazgazan/bacom/releases)
- extract the archive and place the `bacom` binary in your `$PATH`

### From source

- Have go 1.8 or greater installed: [golang.org](https://golang.org/doc/install)
- run `go get -u github.com/yazgazan/bacom/cmd/bacom`

## Usage

Bacom works by comparing a live (locale or testing) version against one or more "known" versions of a service.
These known versions are stored in a `bacom-tests` folder as request/responses pairs.

```text
bacom-tests/
├── config.json
├── v0.0.1
│   ├── api-call_req.txt
│   ├── api-call_resp.txt
│   ├── api-call2_req.txt
│   ├── api-call2_resp.txt
│   ├── api-call3_req.txt
│   └── api-call3_resp.txt
└── v1.0.0
│   ├── api-call_req.txt
│   ├── api-call_resp.txt
│   ├── api-call2_req.txt
│   ├── api-call2_resp.txt
│   ├── api-call3_req.txt
│   ├── api-call3_resp.txt
│   ├── api-call4_req.txt
│   └── api-call4_resp.txt
```

### Importing requests and responses

Requests and responses can be imported from two formats: har and curl.

The HAR format can be used to import requests from google-chrome and firefox,
by exporting one or all requests/responses from the network tab.
Important to note is that the response bodies won't be present when exporting all requests from google-chrome
(these can be generated later via the `bacom test` command).

Importing from HAR files:

```bash
bacom import har -out=bacom-tests/v0.0.1 har_files/*.har
```

The curl import format allows you to import a request using the same command-line options as `curl`:

```bash
bacom import curl -X POST -H "Content-Type: application/json" -d '{"foo": ["bar"]}' "http://localhost:8080/api/endpoint" -dir=bacom-tests/v0.0.1 -name="post-api-endpoint"
```

The curl import can be used to import requests from many sources: google-chrome, firefox, postman, etc.

### Testing a new version

When testing a new version, bacom will replay the requests from older versions against a live endpoint.
A diff will be generated for the request's status code, headers and JSON body.

When comparing the bodies, two kind of errors will be reported:

- Invalid type: the type of a JSON key changed in the new version.
- Missing key: a key is absent in the new version.

Differences in content are not reported.

Testing against versions up (but excluding) `v2.0.0`:

```bash
bacom test -version="<=v1.x" -target-host=localhost:8080
```

A configuration file can be used to specify headers and JSON paths to ignore in the diff:

```bash
bacom test -conf=bacom-ignore.json -version="<=v1.x" -target-host=localhost:8080
```

### Saving responses for a new version

Once a new version is fixed (considered correct), requests and responses can be generated based on the old versions requests:

```bash
bacom test -version="<=v1.x" -target-host=localhost:8080 -save=v2.0.0
```

## Planned features

- [ ] Supporting HTTP trailers
- [ ] Supporting more import formats (Postman and Insomnia)
- [ ] Allow the use of pre-processing commands for requests (i.e: setting authentication headers)
