package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// MakeHTTPRequest makes an HTTP(S) request to the specified URL.
// It returns the HTTP response or an error if the request fails.
func MakeHTTPRequest(url, certPath, keyPath, caPath string, insecure bool, timeout int) (*http.Response, error) {
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

	// Make the HTTP or HTTPS request
	resp, err := client.Get(url)
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
