package tidb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
)

// TestPDGetMemberNames tests the PDGetMemberNames function.
func TestPDGetMemberNames(t *testing.T) {
	// Mock PD server response
	mockResponse := `{
        "members": [
            {"name": "pd-1"},
            {"name": "pd-2"},
            {"name": "pd-3"}
        ]
    }`

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// Create a mock configuration
	conf := cfg.AppConfig{
		PDAddress:          server.URL[len("http://"):], // Remove "http://" prefix
		TLSCertPath:        "",
		TLSKeyPath:         "",
		TLSCAPath:          "",
		TLSInsecure:        false,
		HTTPRequestTimeout: 5,
	}

	// Call the function
	names, err := PDGetMemberNames(conf)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Validate the result
	expected := []string{"pd-1", "pd-2", "pd-3"}
	if len(names) != len(expected) {
		t.Fatalf("Expected %d names, got %d", len(expected), len(names))
	}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Expected name %s, got %s", expected[i], name)
		}
	}
}

// TestGetUniqueDomains tests the GetUniqueDomains function.
func TestGetUniqueDomains(t *testing.T) {
	hosts := []string{"pd-1.example.com", "pd-2.example.com", "pd-3.example.org"}
	expected := []string{"example.com", "example.org"}

	result := GetUniqueDomains(hosts)
	if len(result) != len(expected) {
		t.Fatalf("Expected %d unique domains, got %d", len(expected), len(result))
	}
	for i, domain := range result {
		if domain != expected[i] {
			t.Errorf("Expected domain %s, got %s", expected[i], domain)
		}
	}
}

// TestBuildPDAssistantHostnames tests the BuildPDAssistantHostnames function.
func TestBuildPDAssistantHostnames(t *testing.T) {
	conf := cfg.AppConfig{
		PDAssistantHostPrefix: "pd-assistant",
	}
	pdNames := []string{"pd-1.example.com", "pd-2.example.com", "pd-3.example.org"}
	expected := []string{"pd-assistant.example.com", "pd-assistant.example.org"}

	result := BuildPDAssistantHostnames(conf, pdNames)
	if len(result) != len(expected) {
		t.Fatalf("Expected %d hostnames, got %d", len(expected), len(result))
	}
	for i, hostname := range result {
		if hostname != expected[i] {
			t.Errorf("Expected hostname %s, got %s", expected[i], hostname)
		}
	}
}
