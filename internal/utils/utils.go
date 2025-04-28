package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

// MakeHTTPRequest makes an HTTP(S) request to the specified URL.
// It returns the HTTP response or an error if the request fails.
func MakeHTTPRequest(url, certPath, keyPath, caPath string, insecure bool, timeout int, bearerToken string) (*http.Response, error) {
	var client *http.Client
	var tlsConfig *tls.Config
	var cert tls.Certificate
	var caCertPool *x509.CertPool
	var caCert []byte
	var err error

	if strings.HasPrefix(url, "https:") {
		if certPath != "" && keyPath != "" {
			// Load the client certificate
			cert, err = tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return nil, fmt.Errorf("could not load client certificate: %v", err)
			}
		}

		if caPath != "" {
			// Create a CA certificate pool and add the CA certificate
			caCert, err = os.ReadFile(caPath)
			if err != nil {
				return nil, fmt.Errorf("could not read CA certificate: %v", err)
			}

			caCertPool = x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				return nil, fmt.Errorf("could not append CA certificate to pool")
			}

			// Create the TLS configuration with the client certificate and CA pool
			tlsConfig = &tls.Config{
				Certificates:       []tls.Certificate{cert},
				ClientCAs:          caCertPool,
				InsecureSkipVerify: insecure, // Set to true to skip server verification (not recommended)
			}
		}

		// Create the custom transport
		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		// Create an HTTP client using the custom transport
		client = &http.Client{
			Transport: transport,
			Timeout:   time.Duration(timeout) * time.Second,
		}
	} else {
		// Use a default HTTP client if no CA path is provided
		client = &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	}

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create HTTP request: %v", err)
	}

	// Add the bearer token to the Authorization header if provided
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	// Make the HTTP or HTTPS request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make HTTP(S) request: %v", err)
	}

	return resp, nil
}

// ParseCommaSeparatedLine takes a string with comma-separated values and returns a slice of strings.
func ParseCommaSeparatedLine(line string) []string {
	// Split the line by commas and trim any surrounding whitespace
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// GetDomainFromHost extracts the domain from a given host string.
func GetDomainFromHost(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		return ""
	}
	return strings.Join(parts[1:], ".")
}

// Contains checks if a slice contains a specific string.
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ParseJSONResponse parses a JSON response body into the provided target interface.
func ParseJSONResponse(body io.Reader, target interface{}) error {
	decoder := json.NewDecoder(body)
	return decoder.Decode(target)
}

// IPListsEqual checks if two slices of IPs are equal (uses sort to ensure order doesn't matter).
func IPListsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	slices.Sort(a)
	slices.Sort(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
