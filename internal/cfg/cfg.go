package cfg

import (
	"fmt"
	"os"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/impossiblecloud/pd-cert-assistant/internal/utils"
	"sigs.k8s.io/yaml"
)

// TLSConfig holds the TLS configuration parameters.
type TLSConfig struct {
	CertPath string
	KeyPath  string
	CAPath   string
	Insecure bool
}

// PDConfig holds the configuration parameters for the PD endpoint.
type PDConfig struct {
	Address            string
	TLSConfig          TLSConfig
	HTTPRequestTimeout int
}

// PDDiscoveryConfig holds the configuration parameters for the PD Discovery endpoint.
type PDDiscoveryConfig struct {
	URL                  string
	TLSConfig            TLSConfig
	HTTPRequestTimeout   int
	TiDBCLusterName      string
	TiDBCLusterNameSpace string
}

// AppConfig is the main configuration structure for the application.
type AppConfig struct {
	// PDConfig for pulling data from PD instance.
	PDConfig PDConfig
	// PDDiscoveryConfig is the URL for PD discovery service.
	PDDiscoveryConfig PDDiscoveryConfig
	// BearerToken is the token used for authentication
	BearerToken string
	// CertificateFilePath is the path to the certificate file.
	Certificate cmapi.Certificate

	// PD Assistants host parameters
	PDAssistantURLs        []string
	PDAssistantHostPrefix  string
	PDAssistantScheme      string
	PDAssistantPort        string
	PDAssistantTLSInsecure bool

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
	config.PDConfig = PDConfig{}
	config.PDDiscoveryConfig = PDDiscoveryConfig{}
	// TODO: make timeouts configurable
	config.HTTPRequestTimeout = 5                   // seconds
	config.PDConfig.HTTPRequestTimeout = 5          // seconds
	config.PDDiscoveryConfig.HTTPRequestTimeout = 5 // seconds
	return config
}

// Update updates the AppConfig instance with values from command line arguments and environment variables.
func (c *AppConfig) Update(pdAssistantURLs, certPath string) error {
	// Update config based on command line arguments
	if pdAssistantURLs != "" {
		c.PDAssistantURLs = utils.ParseCommaSeparatedLine(pdAssistantURLs)
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

// Validate checks if the AppConfig instance has valid values.
func (c *AppConfig) Validate() error {
	if c.PDDiscoveryConfig.URL != "" {
		// In this case we require tidb cluster name and namespace
		if c.PDDiscoveryConfig.TiDBCLusterName == "" {
			return fmt.Errorf("PD discovery service requires a TiDB cluster name")
		}
		if c.PDDiscoveryConfig.TiDBCLusterNameSpace == "" {
			return fmt.Errorf("PD discovery service requires a TiDB cluster namespace")
		}
	}
	return nil
}
