package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
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
		return errors.Errorf("unknown version %q. Available versions: v0, v1, v2", s)
	case "0", "v0":
		*v = v0
	case "1", "v1":
		*v = v1
	case "2", "v2":
		*v = v2
	}

	return nil
}

// ResponseV0 is the baseline version of this test server
type ResponseV0 struct {
	Results []V0
}

// V0 is part of the ResponseV0 structure
type V0 struct {
	Foo string
	Bar int
}

// ResponseV1 is a backward-compatible version of this test server
type ResponseV1 struct {
	Results []V1
}

// V1 is part of the ResponseV1 structure
type V1 struct {
	Foo  string
	Bar  int
	Buzz float64
}

// ResponseV2 is a backward-compatibility breaking version of this test server
type ResponseV2 struct {
	Results []V2
}

// V2 is part of the ResponseV2 structure
type V2 struct {
	Bar  []int
	Buzz float64
}

func setV0Headers(h http.Header) {
	h.Set("Content-Type", "application/json")
}

func setV1Headers(h http.Header) {
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", "no-cache")
}

func setV2Headers(h http.Header) {
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", "max-age=300")
}

func v0Handler(stream bool) func(w http.ResponseWriter, req *http.Request) {
	encode := encodeV0
	if stream {
		encode = encodeV0Stream
	}

	return func(w http.ResponseWriter, req *http.Request) {
		setV0Headers(w.Header())

		defer func() {
			err := req.Body.Close()
			if err != nil {
				log.Print("Error:", err)
			}
		}()

		err := encode(w, ResponseV0{
			Results: []V0{
				{
					Foo: "bar",
					Bar: 42,
				},
				{
					Foo: "hello world",
					Bar: 11,
				},
			},
		})

		if err != nil {
			log.Print("error encoding v0 response:", err)
		}
	}
}

func encodeV0(w io.Writer, resp ResponseV0) error {
	return json.NewEncoder(w).Encode(resp)
}

func encodeV0Stream(w io.Writer, resp ResponseV0) error {
	enc := json.NewEncoder(w)
	for _, result := range resp.Results {
		err := enc.Encode(result)
		if err != nil {
			return err
		}
	}

	return nil
}

func v1Handler(stream bool) func(w http.ResponseWriter, req *http.Request) {
	encode := encodeV1
	if stream {
		encode = encodeV1Stream
	}
	return func(w http.ResponseWriter, req *http.Request) {
		setV1Headers(w.Header())

		defer func() {
			err := req.Body.Close()
			if err != nil {
				log.Print("Error:", err)
			}
		}()

		err := encode(w, ResponseV1{
			Results: []V1{
				{
					Foo:  "hello world",
					Bar:  23,
					Buzz: 1.2,
				},
			},
		})

		if err != nil {
			log.Print("error encoding v1 response:", err)
		}
	}
}

func encodeV1(w io.Writer, resp ResponseV1) error {
	return json.NewEncoder(w).Encode(resp)
}

func encodeV1Stream(w io.Writer, resp ResponseV1) error {
	enc := json.NewEncoder(w)
	for _, result := range resp.Results {
		err := enc.Encode(result)
		if err != nil {
			return err
		}
	}

	return nil
}

func v2Handler(stream bool) func(w http.ResponseWriter, req *http.Request) {
	encode := encodeV2
	if stream {
		encode = encodeV2Stream
	}
	return func(w http.ResponseWriter, req *http.Request) {
		setV2Headers(w.Header())

		defer func() {
			err := req.Body.Close()
			if err != nil {
				log.Print("Error:", err)
			}
		}()

		err := encode(w, ResponseV2{
			Results: []V2{
				{
					Bar:  []int{1, 2, 3},
					Buzz: 2.1,
				},
				{
					Bar:  []int{4, 5},
					Buzz: 6.4,
				},
				{
					Bar:  []int{6},
					Buzz: 0.02,
				},
			},
		})

		if err != nil {
			log.Print("error encoding v2 response:", err)
		}
	}
}

func encodeV2(w io.Writer, resp ResponseV2) error {
	return json.NewEncoder(w).Encode(resp)
}

func encodeV2Stream(w io.Writer, resp ResponseV2) error {
	enc := json.NewEncoder(w)
	for _, result := range resp.Results {
		err := enc.Encode(result)
		if err != nil {
			return err
		}
	}

	return nil
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}

func authHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer foo" {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func main() {
	var listen string
	var stream bool
	v := v0
	mux := &http.ServeMux{}

	flag.StringVar(&listen, "listen", "localhost:1235", "host:port to listen on")
	flag.BoolVar(&stream, "stream", false, "return JSON stream")
	flag.Var(&v, "version", "version of the server to run (v0, v1 or v2)")
	flag.Parse()

	srv := &http.Server{
		Addr:    listen,
		Handler: mux,
	}
	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			err := srv.Shutdown(context.Background())
			if err != nil {
				log.Fatal("Error:", err)
			}
			os.Exit(0)
		}()
	})

	switch v {
	default:
		log.Fatalf("unknown version %q", v)
	case v0:
		mux.Handle("/api", authHandler(http.HandlerFunc(v0Handler(stream))))
	case v1:
		mux.Handle("/api", authHandler(http.HandlerFunc(v1Handler(stream))))
	case v2:
		mux.Handle("/api", authHandler(http.HandlerFunc(v2Handler(stream))))
		mux.HandleFunc("/not-found", notFoundHandler)
	}

	log.Printf("running version %s of the server", v)
	log.Printf("listening on %s", listen)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
