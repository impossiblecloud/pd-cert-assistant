package cfg

import (
	"fmt"
	"os"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/impossiblecloud/pd-cert-assistant/internal/utils"
	"sigs.k8s.io/yaml"
)

// AppConfig is the main configuration structure for the application.
type AppConfig struct {
	// PDAddress is the address of the PD server.
	PDAddress string
	// BearerToken is the token used for authentication
	BearerToken string
	// CertificateFilePath is the path to the certificate file.
	Certificate cmapi.Certificate

	// PD Assistants host parameters
	PDAssistantAddresses   []string
	PDAssistantHostPrefix  string
	PDAssistantScheme      string
	PDAssistantPort        string
	PDAssistantTLSInsecure bool

	// TLS Parameters
	TLSCertPath string
	TLSKeyPath  string
	TLSCAPath   string
	TLSInsecure bool

	// HTTPRequestTimeout is the timeout for HTTP requests in seconds.
	HTTPRequestTimeout int
	// KubernetesPollInterval is the interval for polling Kubernetes in seconds.
	KubernetesPollInterval int
	// PDAssistantPollInterval is the interval for polling all pd-assistants in seconds.
	PDAssistantPollInterval int
	// CertUpdateInterval is the interval for updating PD certificate in seconds.
	CertUpdateInterval int
}

// LoadCertificateYaml loads a certificate YAML file and unmarshals it into a Certificate object.
func LoadCertificateYaml(certificateFilePath string) (cmapi.Certificate, error) {
	newCert := cmapi.Certificate{}

	// Load the Certificate YAML file
	certData, err := os.ReadFile(certificateFilePath)
	if err != nil {
		return newCert, fmt.Errorf("failed to read certificate file %s: %s", certificateFilePath, err.Error())
	}

	if err := yaml.Unmarshal(certData, &newCert); err != nil {
		return newCert, fmt.Errorf("failed to unmarshal certificate YAML: %s", err.Error())
	}

	return newCert, nil
}

// Create returns a new AppConfig instance with default values.
func Create() AppConfig {
	config := AppConfig{}
	config.HTTPRequestTimeout = 5 // seconds
	return config
}

// Update updates the AppConfig instance with values from command line arguments and environment variables.
func (c *AppConfig) Update(pdAssistantAddresses, certPath string) error {
	// Update config based on command line arguments
	if pdAssistantAddresses != "" {
		c.PDAssistantAddresses = utils.ParseCommaSeparatedLine(pdAssistantAddresses)
	}

	// Update config with environment variables
	c.BearerToken = os.Getenv("BEARER_TOKEN")
	if c.BearerToken == "" {
		return fmt.Errorf("BEARER_TOKEN environment variable is not set")
	}

	// Load Certificate YAML
	newCert, err := LoadCertificateYaml(certPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate YAML: %s", err.Error())
	}
	c.Certificate = newCert

	return nil
}
