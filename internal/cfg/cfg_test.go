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
