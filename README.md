# apiffuf

API URL fuzzer that cross-joins hosts and paths into normalized URLs, probes them over HTTP, and reports responding endpoints.

<a href="https://www.producthunt.com/products/apiffuf-by-jsmon?embed=true&amp;utm_source=badge-featured&amp;utm_medium=badge&amp;utm_campaign=badge-apiffuf-by-jsmon" target="_blank" rel="noopener noreferrer"><img alt="Apiffuf by Jsmon - API URL fuzzer for API hackers | Product Hunt" width="250" height="54" src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=1154673&amp;theme=light&amp;t=1779692501333"></a>

## Installation

### Direct install

```bash
go install github.com/jsmonhq/apiffuf@latest
```

### Clone and build

```bash
git clone https://github.com/jsmonhq/apiffuf.git
cd apiffuf
go build -ldflags="-s -w" -o apiffuf .
```

## Usage

```text
apiffuf -hosts <host|file> -paths <file> [options]
```

### Flags

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-hosts` | `-u` | — | Host or file containing hosts (required) |
| `-paths` | `-w` | — | File containing API paths (required) |
| `-method` | `-X` | `GET` | HTTP method (supports custom methods) |
| `-headers` | `-H` | — | Request header (`Name: value`, repeatable) |
| `-threads` | `-t` | `20` | Parallel goroutines |
| `-rate` | — | `0` | Requests per second (`0` = unlimited) |
| `-o` | — | — | Save default text output to file |
| `-oJ` | — | — | Save JSON output to file |
| `-oC` | — | — | Save CSV output to file |
| `-timeout` | — | `10s` | Per-request timeout |
| `-user-agent` | — | `apiffuf/1.0` | User-Agent header |
| `-no-color` | — | `false` | Disable colored terminal output |

### Examples

Single host and paths file:

```bash
apiffuf -hosts api.jsmon.sh -paths paths.txt
```

Hosts file and custom method:

```bash
apiffuf -u hosts.txt -w paths.txt -X POST
```

With headers, concurrency, and rate limit:

```bash
apiffuf -hosts https://api.example.com -paths paths.txt -H "Authorization: Bearer token" -t 50 -rate 10
```

Save results:

```bash
apiffuf -hosts api.jsmon.sh -paths paths.txt -o results.txt -oJ results.json -oC results.csv
```

## URL normalization

`apiffuf` normalizes host/path combinations before probing:

| Host | Path | Output |
|------|------|--------|
| `http://sub.target.com` | `/api/v2/users` | `http://sub.target.com/api/v2/users` |
| `http://sub.target.com/` | `/api/v2/users` | `http://sub.target.com/api/v2/users` |
| `http://sub.target.com` | `api/v2/users` | `http://sub.target.com/api/v2/users` |
| `sub.target.com` | `/api/v2/users` | `https://sub.target.com/api/v2/users` |

If no protocol is supplied in the host input, `https` is used by default.

## Output

Default terminal output (colored when stdout is a TTY):

```text
https://api.jsmon.sh/api/v2/users [200] [application/json] [12234] [Jsmon API]
```

Each line includes:

1. URL
2. Status code
3. Content-Type
4. Content-Length
5. Page title (when available)

Only URLs that receive an HTTP response are shown. Connection errors, timeouts, and DNS failures are excluded.

JSON output (`-oJ`) and CSV output (`-oC`) are also supported.

## Safety notice

When using `PUT`, `PATCH`, or `DELETE`, apiffuf prints a caution warning because these methods can modify or delete data. Only use against targets you are authorized to test.

### Built by team [Jsmon](https://jsmon.sh).

## License

AGPLv3
