package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const maxBodyRead = 512 * 1024

// Config holds probe runtime settings.
type Config struct {
	Method     string
	Headers    http.Header
	Timeout    time.Duration
	Threads    int
	Rate       int
	UserAgent  string
	MaxRedirects int
}

// Result holds response metadata for a probed URL.
type Result struct {
	URL           string `json:"url"`
	StatusCode    int    `json:"status_code"`
	ContentType   string `json:"content_type"`
	ContentLength int64  `json:"content_length"`
	Title         string `json:"title"`
	Error         string `json:"error,omitempty"`
}

// Probe issues HTTP requests to urls using concurrent workers.
func Probe(ctx context.Context, urls []string, cfg Config) ([]Result, error) {
	if cfg.Method == "" {
		cfg.Method = http.MethodGet
	}
	if cfg.Threads < 1 {
		return nil, fmt.Errorf("threads must be >= 1")
	}
	if cfg.Rate < 0 {
		return nil, fmt.Errorf("rate must be >= 0")
	}
	if cfg.MaxRedirects <= 0 {
		cfg.MaxRedirects = 10
	}

	client := &http.Client{
		Timeout: cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", cfg.MaxRedirects)
			}
			return nil
		},
	}

	limiter := NewLimiter(cfg.Rate)
	defer limiter.Stop()

	jobs := make(chan string)
	results := make(chan Result, len(urls))

	var wg sync.WaitGroup
	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for urlStr := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				if err := limiter.Wait(ctx); err != nil {
					return
				}
				results <- fetch(ctx, client, urlStr, cfg)
			}
		}()
	}

	go func() {
		for _, u := range urls {
			select {
			case <-ctx.Done():
				break
			case jobs <- u:
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var out []Result
	for r := range results {
		if r.Error == "" {
			out = append(out, r)
		}
	}
	return out, nil
}

func fetch(ctx context.Context, client *http.Client, urlStr string, cfg Config) Result {
	result := Result{URL: urlStr}

	reqCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, cfg.Method, urlStr, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	for key, values := range cfg.Headers {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}
	if req.Header.Get("User-Agent") == "" && cfg.UserAgent != "" {
		req.Header.Set("User-Agent", cfg.UserAgent)
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ContentType = contentType(resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyRead))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if cl := resp.ContentLength; cl >= 0 {
		result.ContentLength = cl
	} else {
		result.ContentLength = int64(len(body))
	}

	result.Title = extractTitle(body)
	return result
}

func contentType(raw string) string {
	if raw == "" {
		return ""
	}
	if idx := strings.Index(raw, ";"); idx >= 0 {
		return strings.TrimSpace(raw[:idx])
	}
	return strings.TrimSpace(raw)
}

func extractTitle(body []byte) string {
	z := html.NewTokenizer(strings.NewReader(string(body)))
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return ""
		case html.StartTagToken, html.SelfClosingTagToken:
			if z.Token().Data == "title" {
				tt = z.Next()
				if tt == html.TextToken {
					return strings.TrimSpace(z.Token().Data)
				}
			}
		}
	}
}
