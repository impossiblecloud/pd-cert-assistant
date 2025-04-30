package cfg

import (
	"testing"
)

func TestLoadCertificateYaml(t *testing.T) {
	cert, err := LoadCertificateYaml("../../fixtures/certificate.yaml")
	if err != nil {
		t.Errorf("failed to load certificate YAML: %v", err)
	}

	if cert.Kind != "Certificate" {
		t.Errorf("expected kind 'Certificate', got '%s'", cert.Kind)
	}
	if cert.APIVersion != "cert-manager.io/v1" {
		t.Errorf("expected API version 'cert-manager.io/v1', got '%s'", cert.APIVersion)
	}
	if cert.Name != "example-certificate" {
		t.Errorf("expected name 'example-certificate', got '%s'", cert.Name)
	}
	if cert.Namespace != "default" {
		t.Errorf("expected namespace 'default', got '%s'", cert.Namespace)
	}
	if len(cert.Spec.DNSNames) != 4 {
		t.Errorf("expected 4 DNS names, got %d", len(cert.Spec.DNSNames))
	}
	if cert.Spec.SecretName != "example-certificate-secret" {
		t.Errorf("expected secret name 'example-certificate-secret', got '%s'", cert.Spec.SecretName)
	}
}
func TestCreate(t *testing.T) {
	config := Create()

	// Check default HTTPRequestTimeout
	if config.HTTPRequestTimeout != 5 {
		t.Errorf("expected HTTPRequestTimeout to be 5, got %d", config.HTTPRequestTimeout)
	}

	// Check PDConfig HTTPRequestTimeout
	if config.PDConfig.HTTPRequestTimeout != 5 {
		t.Errorf("expected PDConfig.HTTPRequestTimeout to be 5, got %d", config.PDConfig.HTTPRequestTimeout)
	}

	// Check PDDiscoveryConfig HTTPRequestTimeout
	if config.PDDiscoveryConfig.HTTPRequestTimeout != 5 {
		t.Errorf("expected PDDiscoveryConfig.HTTPRequestTimeout to be 5, got %d", config.PDDiscoveryConfig.HTTPRequestTimeout)
	}

	// Check if PDConfig is initialized
	if config.PDConfig == (PDConfig{}) {
		t.Errorf("expected PDConfig to be initialized, got an empty struct")
	}

	// Check if PDDiscoveryConfig is initialized
	if config.PDDiscoveryConfig == (PDDiscoveryConfig{}) {
		t.Errorf("expected PDDiscoveryConfig to be initialized, got an empty struct")
	}
}
