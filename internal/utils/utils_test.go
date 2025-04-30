package utils

import (
	"testing"
)

// TestParseCommaSeparatedLine tests the ParseCommaSeparatedLine function.
func TestParseCommaSeparatedLine(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"param1,param2,param3", []string{"param1", "param2", "param3"}},
		{" param1 , param2 , param3 ", []string{"param1", "param2", "param3"}},
		{"", []string{""}},
		{"param1", []string{"param1"}},
	}

	for _, test := range tests {
		result := ParseCommaSeparatedLine(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("Expected length %d, got %d", len(test.expected), len(result))
		}
		for i := range result {
			if result[i] != test.expected[i] {
				t.Errorf("Expected %s, got %s", test.expected[i], result[i])
			}
		}
	}
}

func TestParseCommaSeparatedLineWithEmptyValues(t *testing.T) {
	input := "param1,,param3"
	expected := []string{"param1", "", "param3"}
	result := ParseCommaSeparatedLine(input)
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], result[i])
		}
	}
}

func TestParseCommaSeparatedLineWithSpecialCharacters(t *testing.T) {
	input := "param1,param2,param3!@#$%^&*()"
	expected := []string{"param1", "param2", "param3!@#$%^&*()"}
	result := ParseCommaSeparatedLine(input)
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], result[i])
		}
	}
}

// TestMakeHTTPSRequest tests the MakeHTTPSRequest function.
func TestMakeHTTPRequestWithInvalidURL(t *testing.T) {
	_, err := MakeHTTPRequest(":", "", "", "", false, 2, "")
	if err == nil {
		t.Errorf("Expected error due to invalid URL, got nil")
	}
}

func TestMakeHTTPRequestWithoutCerts(t *testing.T) {
	// Test with a valid URL but without certificates
	_, err := MakeHTTPRequest("https://example.com", "", "", "", false, 2, "")
	if err != nil {
		t.Errorf("Expected no error for request without certificates, got: %v", err)
	}
}

func TestMakeHTTPRequestWithInsecureSkipVerify(t *testing.T) {
	// Test with insecure skip verify set to true
	_, err := MakeHTTPRequest("https://example.com", "", "", "", true, 2, "")
	if err != nil {
		t.Errorf("Expected no error for insecure request, got: %v", err)
	}
}

// TestContains tests the Contains function.
func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"apple", "banana", "cherry"}, "banana", true},
		{[]string{"apple", "banana", "cherry"}, "grape", false},
		{[]string{}, "banana", false},
		{[]string{"apple", "banana", "banana"}, "banana", true},
		{[]string{"apple", "banana", "cherry"}, "", false},
		{[]string{"", "banana", "cherry"}, "", true},
	}

	for _, test := range tests {
		result := Contains(test.slice, test.item)
		if result != test.expected {
			t.Errorf("For slice %v and item %q, expected %v, got %v", test.slice, test.item, test.expected, result)
		}
	}
}

// TestGetDomainFromHost tests the GetDomainFromHost function.
func TestGetDomainFromHost(t *testing.T) {
	tests := []struct {
		host     string
		expected string
	}{
		{"example.com", ""},
		{"sub.example.com", "example.com"},
		{"sub.sub.example.com", "sub.example.com"},
		{"localhost", ""},
		{"", ""},
		{"example", ""},
		{"pd.namespace.svc.cluster.local", "namespace.svc.cluster.local"},
	}

	for _, test := range tests {
		result := GetDomainFromHost(test.host)
		if result != test.expected {
			t.Errorf("For host %q, expected %q, got %q", test.host, test.expected, result)
		}
	}
}

// TestIPListsEqual tests the IPListsEqual function.
func TestIPListsEqual(t *testing.T) {
	tests := []struct {
		a, b     []string
		expected bool
	}{
		{[]string{"192.168.1.1", "10.0.0.1"}, []string{"10.0.0.1", "192.168.1.1"}, true},
		{[]string{"192.168.1.1", "10.0.0.1"}, []string{"192.168.1.1", "10.0.0.1"}, true},
		{[]string{"192.168.1.1", "10.0.0.1"}, []string{"192.168.1.1", "10.0.0.2"}, false},
		{[]string{"192.168.1.1"}, []string{"192.168.1.1"}, true},
		{[]string{"192.168.1.1"}, []string{"10.0.0.1"}, false},
		{[]string{}, []string{}, true},
		{[]string{"192.168.1.1", "192.168.1.1"}, []string{"192.168.1.1"}, false},
		{[]string{"192.168.1.1"}, []string{"192.168.1.1", "192.168.1.1"}, false},
	}

	for _, test := range tests {
		result := IPListsEqual(test.a, test.b)
		if result != test.expected {
			t.Errorf("For lists %v and %v, expected %v, got %v", test.a, test.b, test.expected, result)
		}
	}
}

// TestFindUniqueURLs tests the FindUniqueURLs function.
func TestFindUniqueURLs(t *testing.T) {
	tests := []struct {
		text     string
		expected []string
	}{
		{
			text:     "http://example.com\nhttps://example.org\nhttp://example.com",
			expected: []string{"http://example.com", "https://example.org"},
		},
		{
			text:     "Visit http://example-1.com and https://example-2.org for more info.",
			expected: []string{"http://example-1.com", "https://example-2.org"},
		},
		{
			text:     "No URLs here.",
			expected: []string{},
		},
		{
			text:     "",
			expected: []string{},
		},
		{
			text: "--join=https://tidb-cluster-pd-1.tidb-cluster-pd-peer.tidb-regional.svc.de-fra002.mydomain.com:2379,https://tidb-cluster-pd-1.tidb-cluster-pd-peer.tidb-regional.svc.de-fra001.mydomain.com:2379,https://tidb-cluster-pd-0.tidb-cluster-pd-peer.tidb-regional.svc.de-fra002.mydomain.com:2379,https://eu-central-2-pd-0.eu-central-2-pd-peer.tidb-regional.svc.nl-ams001.mydomain.com:2379,https://tidb-cluster-pd-0.tidb-cluster-pd-peer.tidb-regional.svc.de-fra001.mydomain.com:2379",
			expected: []string{
				"https://tidb-cluster-pd-1.tidb-cluster-pd-peer.tidb-regional.svc.de-fra002.mydomain.com:2379",
				"https://tidb-cluster-pd-1.tidb-cluster-pd-peer.tidb-regional.svc.de-fra001.mydomain.com:2379",
				"https://tidb-cluster-pd-0.tidb-cluster-pd-peer.tidb-regional.svc.de-fra002.mydomain.com:2379",
				"https://eu-central-2-pd-0.eu-central-2-pd-peer.tidb-regional.svc.nl-ams001.mydomain.com:2379",
				"https://tidb-cluster-pd-0.tidb-cluster-pd-peer.tidb-regional.svc.de-fra001.mydomain.com:2379",
			},
		},
		{
			text:     "--initial-cluster=tidb-cluster-pd-0.tidb-cluster-pd-peer.tidb-test.svc=https://tidb-cluster-pd-0.tidb-cluster-pd-peer.tidb-test.svc:2380",
			expected: []string{"https://tidb-cluster-pd-0.tidb-cluster-pd-peer.tidb-test.svc:2380"},
		},
		{
			text: "--initial-cluster infra0=http://10.0.1.10:2380,infra1=http://10.0.1.11:2380,infra2=http://10.0.1.12:2380",
			expected: []string{
				"http://10.0.1.10:2380",
				"http://10.0.1.11:2380",
				"http://10.0.1.12:2380",
			},
		},
	}

	for _, test := range tests {
		result := FindUniqueURLs(test.text)
		if len(result) != len(test.expected) {
			t.Errorf("For text %q, expected %v, got %v", test.text, test.expected, result)
		}
		for _, url := range test.expected {
			found := false
			for _, resURL := range result {
				if url == resURL {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected URL %q not found in result %v", url, result)
			}
		}
	}
}

// TestGetHostFromURL tests the GetHostFromURL function.
func TestGetHostFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"http://example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"http://example.com:8080", "example.com"},
		{"https://example.com:443", "example.com"},
		{"http://example.com/path/to/resource", "example.com"},
		{"https://example.com:8080/path/to/resource", "example.com"},
		{"example.com", "example.com"},
		{"example.com:8080", "example.com"},
		{"http://localhost", "localhost"},
		{"https://localhost:8443", "localhost"},
		{"http://127.0.0.1", "127.0.0.1"},
		{"https://127.0.0.1:8080", "127.0.0.1"},
		{"", ""},
		{"http://", ""},
		{"https://", ""},
	}

	for _, test := range tests {
		result := GetHostFromURL(test.url)
		if result != test.expected {
			t.Errorf("For URL %q, expected %q, got %q", test.url, test.expected, result)
		}
	}
}
