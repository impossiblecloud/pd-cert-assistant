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
	_, err := MakeHTTPRequest(":", "", "", "", false, 2)
	if err == nil {
		t.Errorf("Expected error due to invalid URL, got nil")
	}
}

func TestMakeHTTPRequestWithoutCerts(t *testing.T) {
	// Test with a valid URL but without certificates
	_, err := MakeHTTPRequest("https://example.com", "", "", "", false, 2)
	if err != nil {
		t.Errorf("Expected no error for request without certificates, got: %v", err)
	}
}

func TestMakeHTTPRequestWithInsecureSkipVerify(t *testing.T) {
	// Test with insecure skip verify set to true
	_, err := MakeHTTPRequest("https://example.com", "", "", "", true, 2)
	if err != nil {
		t.Errorf("Expected no error for insecure request, got: %v", err)
	}
}
