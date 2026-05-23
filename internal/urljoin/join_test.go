package urljoin

import "testing"

func TestJoin(t *testing.T) {
	tests := []struct {
		host string
		path string
		want string
	}{
		{"http://sub.target.com", "/api/v2/users", "http://sub.target.com/api/v2/users"},
		{"http://sub.target.com/", "/api/v2/users", "http://sub.target.com/api/v2/users"},
		{"http://sub.target.com", "api/v2/users", "http://sub.target.com/api/v2/users"},
		{"sub.target.com", "/api/v2/users", "https://sub.target.com/api/v2/users"},
		{"https://host/", "api/v2/users", "https://host/api/v2/users"},
		{"http://host:8080", "/api", "http://host:8080/api"},
		{"https://host/extra/", "/abs", "https://host/abs"},
		{"https://host/extra/", "rel", "https://host/extra/rel"},
		{"https://host", "/api//v2//users", "https://host/api/v2/users"},
		{"  host  ", "  /api  ", "https://host/api"},
	}

	for _, tt := range tests {
		got, err := Join(tt.host, tt.path)
		if err != nil {
			t.Fatalf("Join(%q, %q) error = %v", tt.host, tt.path, err)
		}
		if got != tt.want {
			t.Fatalf("Join(%q, %q) = %q, want %q", tt.host, tt.path, got, tt.want)
		}
	}
}
