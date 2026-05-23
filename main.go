package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jsmonhq/apiffuf/internal/input"
	"github.com/jsmonhq/apiffuf/internal/output"
	"github.com/jsmonhq/apiffuf/internal/probe"
	"github.com/jsmonhq/apiffuf/internal/urljoin"
)

const usageText = `apiffuf — API URL fuzzer

Usage:
  apiffuf -hosts <host|file> -paths <file> [options]

Required:
  -hosts string
        Host or file containing hosts (one per line)
  -u string
        Alias for -hosts
  -paths string
        File containing API paths (one per line)
  -w string
        Alias for -paths

HTTP:
  -method string
        HTTP method (default "GET")
  -X string
        Alias for -method
  -headers value
        Request header in "Name: value" format (repeatable)
  -H value
        Alias for -headers

Concurrency:
  -threads int
        Number of parallel goroutines (default 20)
  -t int
        Alias for -threads
  -rate int
        Rate limit in requests per second; 0 = unlimited (default 0)

Output:
  -o string
        Save default text output to file
  -oJ string
        Save JSON output to file
  -oC string
        Save CSV output to file

General:
  -timeout duration
        Per-request timeout (default 10s)
  -user-agent string
        User-Agent header (default "apiffuf/1.0")
  -no-color
        Disable colored terminal output

Examples:
  apiffuf -hosts api.jsmon.sh -paths paths.txt
  apiffuf -u hosts.txt -w paths.txt -X POST -H "Authorization: Bearer token"
  apiffuf -hosts https://api.example.com -paths paths.txt -t 50 -rate 10 -o results.txt

Checkout https://github.com/jsmonhq/apiffuf
Built by team Jsmon (https://jsmon.sh)
`

type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ", ")
}

func (h *headerFlags) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
	}

	hosts := flag.String("hosts", "", "host or file containing hosts")
	hostsAlias := flag.String("u", "", "alias for -hosts")
	paths := flag.String("paths", "", "file containing API paths")
	pathsAlias := flag.String("w", "", "alias for -paths")
	method := flag.String("method", "", "HTTP method")
	methodAlias := flag.String("X", "", "alias for -method")
	var headers headerFlags
	flag.Var(&headers, "headers", "request header (repeatable)")
	flag.Var(&headers, "H", "alias for -headers")
	threads := flag.Int("threads", 20, "parallel goroutines")
	threadsAlias := flag.Int("t", 0, "alias for -threads")
	rate := flag.Int("rate", 0, "requests per second (0 = unlimited)")
	outFile := flag.String("o", "", "save default output to file")
	outJSON := flag.String("oJ", "", "save JSON output to file")
	outCSV := flag.String("oC", "", "save CSV output to file")
	timeout := flag.Duration("timeout", 10*time.Second, "per-request timeout")
	userAgent := flag.String("user-agent", "apiffuf/1.0", "User-Agent header")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Parse()

	hostInput := coalesce(*hostsAlias, *hosts)
	pathInput := coalesce(*pathsAlias, *paths)
	methodInput := coalesce(*methodAlias, *method)
	if methodInput == "" {
		methodInput = "GET"
	}
	threadCount := *threads
	if *threadsAlias > 0 {
		threadCount = *threadsAlias
	}

	if hostInput == "" || pathInput == "" {
		fmt.Fprintln(os.Stderr, "error: -hosts and -paths are required")
		flag.Usage()
		os.Exit(1)
	}
	if threadCount < 1 {
		fmt.Fprintln(os.Stderr, "error: -threads must be >= 1")
		os.Exit(1)
	}
	if *rate < 0 {
		fmt.Fprintln(os.Stderr, "error: -rate must be >= 0")
		os.Exit(1)
	}

	headerMap, err := output.ParseHeaders(headers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	hostList, err := input.LoadHosts(hostInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	pathList, err := input.LoadPaths(pathInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	urls, err := buildURLs(hostList, pathList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	warnSensitiveMethod(methodInput)

	httpHeaders := make(http.Header)
	for k, v := range headerMap {
		httpHeaders.Set(k, v)
	}

	cfg := probe.Config{
		Method:    strings.ToUpper(methodInput),
		Headers:   httpHeaders,
		Timeout:   *timeout,
		Threads:   threadCount,
		Rate:      *rate,
		UserAgent: *userAgent,
	}

	ctx := context.Background()
	results, err := probe.Probe(ctx, urls, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	opts := output.Options{Color: output.SupportsColor(os.Stdout, *noColor)}

	if *outFile != "" {
		if err := output.WriteFile(*outFile, results, func(w io.Writer, r []probe.Result) error {
			return output.WriteDefault(w, r, output.Options{Color: false})
		}); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
	if *outJSON != "" {
		if err := output.WriteFile(*outJSON, results, func(w io.Writer, r []probe.Result) error {
			return output.WriteJSON(w, r)
		}); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
	if *outCSV != "" {
		if err := output.WriteFile(*outCSV, results, func(w io.Writer, r []probe.Result) error {
			return output.WriteCSV(w, r)
		}); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	if *outFile == "" && *outJSON == "" && *outCSV == "" {
		if err := output.WriteDefault(os.Stdout, results, opts); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func coalesce(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func buildURLs(hosts, paths []string) ([]string, error) {
	var urls []string
	for _, host := range hosts {
		for _, path := range paths {
			u, err := urljoin.Join(host, path)
			if err != nil {
				return nil, fmt.Errorf("join %q + %q: %w", host, path, err)
			}
			urls = append(urls, u)
		}
	}
	return urls, nil
}

func warnSensitiveMethod(method string) {
	switch strings.ToUpper(method) {
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		fmt.Fprintln(os.Stderr, "[!] CAUTION: PUT/PATCH/DELETE can modify or delete server data. Proceed only on authorized targets.")
	}
}
