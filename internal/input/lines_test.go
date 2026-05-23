package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadLinesSkipsCommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	content := "# comment\n\n  host1  \n# another\nhost2\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines() error = %v", err)
	}
	want := []string{"host1", "host2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestLoadHostsFileAndSingle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.txt")
	if err := os.WriteFile(path, []byte("a.example.com\nb.example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fromFile, err := LoadHosts(path)
	if err != nil {
		t.Fatalf("LoadHosts(file) error = %v", err)
	}
	if len(fromFile) != 2 {
		t.Fatalf("got %d hosts, want 2", len(fromFile))
	}

	fromSingle, err := LoadHosts("api.jsmon.sh")
	if err != nil {
		t.Fatalf("LoadHosts(single) error = %v", err)
	}
	if len(fromSingle) != 1 || fromSingle[0] != "api.jsmon.sh" {
		t.Fatalf("got %v, want [api.jsmon.sh]", fromSingle)
	}
}

func TestLoadPathsRequiresFile(t *testing.T) {
	_, err := LoadPaths("missing.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "paths.txt")
	if err := os.WriteFile(path, []byte("/api/v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadPaths(path)
	if err != nil {
		t.Fatalf("LoadPaths() error = %v", err)
	}
	if len(got) != 1 || got[0] != "/api/v1" {
		t.Fatalf("got %v, want [/api/v1]", got)
	}
}
