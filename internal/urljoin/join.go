package urljoin

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var multiSlash = regexp.MustCompile(`/{2,}`)

// Join normalizes and combines a host/base URL with a path segment.
func Join(host, path string) (string, error) {
	host = strings.TrimSpace(host)
	path = strings.TrimSpace(path)
	if host == "" {
		return "", fmt.Errorf("empty host")
	}
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	if !strings.Contains(host, "://") {
		host = "https://" + host
	}

	base, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("parse host: %w", err)
	}
	if base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("invalid host: %s", host)
	}

	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("parse path: %w", err)
	}

	joined := base.ResolveReference(ref)
	joined.Path = collapseSlashes(joined.Path)
	joined.Fragment = ""
	joined.RawFragment = ""

	return joined.String(), nil
}

func collapseSlashes(path string) string {
	if path == "" {
		return "/"
	}
	collapsed := multiSlash.ReplaceAllString(path, "/")
	if !strings.HasPrefix(collapsed, "/") {
		collapsed = "/" + collapsed
	}
	return collapsed
}
