package cfg

// AppConfig is the main configuration structure for the application.
type AppConfig struct {
	// PDAddress is the address of the PD server.
	PDAddress string
	// PDAssistantHostPrefix is the host prefix for PD Assistant instances.
	PDAssistantHostPrefix string
	// BearerToken is the token used for authentication
	BearerToken string
	// CertificateName is the name of the certificate to be used.
	CertificateName string

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
