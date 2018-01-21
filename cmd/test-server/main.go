package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"flag"
	"github.com/pkg/errors"
	"os"
	"io"
)

type version string

const (
	v0 version = "v0"
	v1 version = "v1"
	v2 version = "v2"
)

func (v version) String() string {
	return string(v)
}

func (v *version) Set(s string) error {
	switch strings.ToLower(s) {
	default:
		errors.Errorf("unknown version %q. Available versions: v0, v1, v2")
	case "0", "v0":
		*v = v0
	case "1", "v1":
		*v = v1
	case "2", "v2":
		*v = v2
	}

	return nil
}

// V0 is the baseline version of this test server
type V0 struct {
	Foo string
	Bar int
}

// V1 is a backward-compatible version of this test server
type V1 struct {
	Foo  string
	Bar  int
	Buzz float64
}

// V2 is a backward-compatibility breaking version of this test server
type V2 struct {
	Foo  string
	Bar  []int
	Buzz float64
}

func setV0Headers(h http.Header) {
	h.Set("Content-Type", "application/json")
}

func setV1Headers(h http.Header) {
	h.Set("Content-Type", "application/json")
}

func setV2Headers(h http.Header) {
	h.Set("Content-Type", "application/json")
}

func v0Handler(w http.ResponseWriter, req *http.Request) {
	setV0Headers(w.Header())

	io.Copy(os.Stdout, req.Body)
	defer req.Body.Close()
	err := json.NewEncoder(w).Encode(V0{
		Foo: "bar",
		Bar: 42,
	})

	if err != nil {
		log.Printf("error encoding v0 response: ", err)
	}
}

func v1Handler(w http.ResponseWriter, req *http.Request) {
	setV1Headers(w.Header())

	io.Copy(os.Stdout, req.Body)
	defer req.Body.Close()

	err := json.NewEncoder(w).Encode(V1{
		Foo:  "hello world",
		Bar:  23,
		Buzz: 1.2,
	})

	if err != nil {
		log.Printf("error encoding v0 response: ", err)
	}
}

func v2Handler(w http.ResponseWriter, req *http.Request) {
	setV2Headers(w.Header())

	io.Copy(os.Stdout, req.Body)
	defer req.Body.Close()

	err := json.NewEncoder(w).Encode(V2{
		Foo:  "bar",
		Bar:  []int{1, 2, 3},
		Buzz: 2.1,
	})

	if err != nil {
		log.Printf("error encoding v0 response: ", err)
	}
}

func main() {
	var listen string
	v := v0
	mux := &http.ServeMux{}

	flag.StringVar(&listen, "listen", "localhost:1235", "host:port to listen on")
	flag.Var(&v, "version", "version of the server to run (v0, v1 or v2)")
	flag.Parse()

	srv := &http.Server{
		Addr:    listen,
		Handler: mux,
	}
	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		os.Exit(0)
	})

	switch v {
	default:
		log.Fatalf("unknown version %q", v)
	case v0:
		mux.HandleFunc("/api", v0Handler)
	case v1:
		mux.HandleFunc("/api", v1Handler)
	case v2:
		mux.HandleFunc("/api", v2Handler)
	}

	log.Printf("running version %s of the server", v)
	log.Printf("listening on %s", listen)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
