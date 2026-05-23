package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/jsmonhq/apiffuf/internal/probe"
)

// Options controls output formatting.
type Options struct {
	Color bool
}

// WriteDefault writes human-readable results to w.
func WriteDefault(w io.Writer, results []probe.Result, opts Options) error {
	sorted := sortResults(results)
	for _, r := range sorted {
		if _, err := fmt.Fprintln(w, formatLine(r, opts.Color)); err != nil {
			return err
		}
	}
	return nil
}

// WriteJSON writes results as a JSON array to w.
func WriteJSON(w io.Writer, results []probe.Result) error {
	sorted := sortResults(results)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(sorted)
}

// WriteCSV writes results as CSV to w.
func WriteCSV(w io.Writer, results []probe.Result) error {
	sorted := sortResults(results)
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"url", "status_code", "content_type", "content_length", "title"}); err != nil {
		return err
	}
	for _, r := range sorted {
		if err := cw.Write([]string{
			r.URL,
			fmt.Sprintf("%d", r.StatusCode),
			r.ContentType,
			fmt.Sprintf("%d", r.ContentLength),
			r.Title,
		}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// WriteFile writes results using write to the given path.
func WriteFile(path string, results []probe.Result, write func(io.Writer, []probe.Result) error) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()
	return write(f, results)
}

func sortResults(results []probe.Result) []probe.Result {
	sorted := make([]probe.Result, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].URL < sorted[j].URL
	})
	return sorted
}

func formatLine(r probe.Result, color bool) string {
	if !color {
		return fmt.Sprintf("%s [%d] [%s] [%d] [%s]", r.URL, r.StatusCode, r.ContentType, r.ContentLength, r.Title)
	}

	statusColor := statusANSI(r.StatusCode)
	return fmt.Sprintf("%s%s%s %s[%d]%s %s[%s]%s %s[%d]%s %s[%s]%s",
		ansiBoldCyan, r.URL, ansiReset,
		statusColor, r.StatusCode, ansiReset,
		ansiDim, r.ContentType, ansiReset,
		ansiDim, r.ContentLength, ansiReset,
		ansiDim, r.Title, ansiReset,
	)
}

func statusANSI(code int) string {
	switch {
	case code >= 200 && code < 300:
		return ansiGreen
	case code >= 300 && code < 400:
		return ansiYellow
	default:
		return ansiRed
	}
}

// SupportsColor reports whether w is a TTY and color is enabled.
func SupportsColor(w io.Writer, noColor bool) bool {
	if noColor {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

const (
	ansiReset    = "\033[0m"
	ansiBoldCyan = "\033[1;36m"
	ansiGreen    = "\033[32m"
	ansiYellow   = "\033[33m"
	ansiRed      = "\033[31m"
	ansiDim      = "\033[2m"
)

// ParseHeaders parses curl-style "Name: value" headers.
func ParseHeaders(raw []string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	headers := make(map[string]string, len(raw))
	for _, h := range raw {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return nil, fmt.Errorf("invalid header format %q, expected \"Name: value\"", h)
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return headers, nil
}
