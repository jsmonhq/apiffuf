package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jsmonhq/apiffuf/internal/probe"
)

func TestWriteCSV(t *testing.T) {
	results := []probe.Result{{
		URL:           "https://example.com",
		StatusCode:    200,
		ContentType:   "text/html",
		ContentLength: 10,
		Title:         "Hello, \"world\"",
	}}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, results); err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "url,status_code,content_type,content_length,title") {
		t.Fatalf("missing header: %q", out)
	}
	if !strings.Contains(out, "\"Hello, \"\"world\"\"\"") {
		t.Fatalf("csv escaping failed: %q", out)
	}
}

func TestWriteJSON(t *testing.T) {
	results := []probe.Result{{
		URL:        "https://example.com",
		StatusCode: 200,
	}}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, results); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	var decoded []probe.Result
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(decoded) != 1 || decoded[0].URL != "https://example.com" {
		t.Fatalf("unexpected json decode: %+v", decoded)
	}
}

func TestParseHeaders(t *testing.T) {
	got, err := ParseHeaders([]string{"Authorization: Bearer token", "X-Test: 1"})
	if err != nil {
		t.Fatalf("ParseHeaders() error = %v", err)
	}
	if got["Authorization"] != "Bearer token" || got["X-Test"] != "1" {
		t.Fatalf("unexpected headers: %+v", got)
	}

	if _, err := ParseHeaders([]string{"bad-header"}); err == nil {
		t.Fatal("expected error for invalid header")
	}
}
