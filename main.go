package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"

	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/k8s"
	"github.com/impossiblecloud/pd-cert-assistant/internal/metrics"
	"github.com/impossiblecloud/pd-cert-assistant/internal/server"
)

// Constants
var Version string

func main() {
	var listen, kubeconfig, pdAssistantURLs, certFilePath string
	var showVersion bool

	if Version == "" {
		Version = "unknown"
	}

	// Init config
	config := cfg.Create()

	// Init state
	srv := server.State{}
	// Init metric
	srv.Metrics = metrics.InitMetrics(Version)

	// General parameters
	flag.StringVar(&listen, "listen", ":8765", "Address:port to listen on")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	// Kubernetes parameters
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file (optional)")
	flag.IntVar(&config.KubernetesPollInterval, "k8s-poll-interval", 60, "Interval for polling Kubernetes in seconds")
	// PD assistant parameters
	flag.IntVar(&config.PDAssistantPollInterval, "pd-assistant-poll-interval", 120, "Interval for polling all pd-assistants and checking/updating certificate, in seconds")
	flag.StringVar(&config.PDAssistantHostPrefix, "pd-assistant-host-prefix", "pd-assistant", "Host prefix for PD Assistant instances")
	flag.StringVar(&config.PDAssistantScheme, "pd-assistant-scheme", "https", "Scheme for PD Assistant instances (http or https)")
	flag.StringVar(&config.PDAssistantPort, "pd-assistant-port", "443", "Port for PD Assistant instances")
	flag.BoolVar(&config.PDAssistantTLSInsecure, "pd-assistant-tls-insecure", false, "Skip TLS verification for PD Assistant instances (not recommended)")
	flag.StringVar(&pdAssistantURLs, "pd-assistant-urls", "", "List of PD Assistant URLs (comma-separated). Overrides --pd-assistant-host-prefix and ignores --pd-address auto-discovery if provided")
	flag.BoolVar(&config.PDAssistantConsensus, "pd-assistant-consensus", false, "Require consensus from all PD Assistant instances before updating the certificate")
	// Certificate parameters
	flag.StringVar(&certFilePath, "certificate-file", "/app/conf/", "Path to a Certificate YAML file to be used as a template")
	// PD discovery parameters
	flag.StringVar(&config.PDDiscoveryConfig.URL, "pd-discovery-url", "", "PD Discovery service URL")
	flag.StringVar(&config.PDDiscoveryConfig.TiDBCLusterName, "pd-discovery-tidb-cluster-name", "", "TiDB cluster name for PD Discovery service")
	flag.StringVar(&config.PDDiscoveryConfig.TiDBCLusterNameSpace, "pd-discovery-tidb-cluster-namespace", "", "TiDB cluster namespace for PD Discovery service")
	flag.Parse()

	// Show and exit functions
	if showVersion {
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	// Update config
	if err := config.Update(pdAssistantURLs, certFilePath); err != nil {
		glog.Fatalf("Failed to update config: %v", err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		glog.Fatalf("Invalid configuration: %v", err)
	}

	// Init k8s client
	kubeClient := k8s.Client{}
	err := kubeClient.Init(kubeconfig)
	if err != nil {
		glog.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Log some useful information
	glog.V(4).Infof("Starting application. Version: %s", Version)
	if len(config.PDAssistantURLs) > 0 {
		glog.V(4).Infof("PD Assistant URLs: %v", config.PDAssistantURLs)
	}
	if len(config.PDDiscoveryConfig.URL) > 0 {
		glog.V(4).Infof("PD Discovery URL: %s", config.PDDiscoveryConfig.URL)
	}
	glog.V(4).Infof("Loaded certificate YAML file %q: name=%s, namespace=%s", certFilePath, config.Certificate.Name, config.Certificate.Namespace)
	if config.PDAssistantConsensus {
		glog.V(4).Infof("PD Assistant consensus check is enabled")
	} else {
		glog.V(4).Infof("PD Assistant consensus check is disabled")
	}

	// Let's rock and roll!
	// Watch CliliumNode IPs and update the state
	go srv.IPWatchLoop(config, kubeClient)

	// Watch all pd-assistant IPs and update the certificate if needed
	go srv.FetchIPsAndUpdateCertLoop(config, kubeClient)

	// Start the main web server
	srv.RunMainWebServer(config, listen)
}
