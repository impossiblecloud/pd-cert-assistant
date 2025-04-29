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

	flag.StringVar(&listen, "listen", ":8765", "Address:port to listen on")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.BoolVar(&config.TLSInsecure, "tls-insecure", false, "Skip TLS verification (not recommended)")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file (optional)")
	flag.IntVar(&config.KubernetesPollInterval, "k8s-poll-interval", 180, "Interval for polling Kubernetes in seconds")
	flag.IntVar(&config.PDAssistantPollInterval, "pd-assistant-poll-interval", 300, "Interval for polling all pd-assistants in seconds")
	flag.IntVar(&config.CertUpdateInterval, "cert-update-interval", 300, "Interval for updating PD certificate in seconds")
	// FIXME: autodiscovery using --pd-address is disabled due to chicken egg problem:
	// 		  pd requires cert to start, pd-assistant can't create cert without pd
	// flag.StringVar(&config.PDAssistantHostPrefix, "pd-assistant-host-prefix", "pd-assistant", "Host prefix for PD Assistant instances")
	// flag.StringVar(&config.PDAssistantScheme, "pd-assistant-scheme", "https", "Scheme for PD Assistant instances (http or https)")
	// flag.StringVar(&config.PDAssistantPort, "pd-assistant-port", "443", "Port for PD Assistant instances")
	// flag.StringVar(&config.PDAddress, "pd-address", "tidb-cluster-pd:2379", "Address:port of PD server")
	// flag.StringVar(&config.TLSCertPath, "tls-cert", "", "Path to the TLS certificate file")
	// flag.StringVar(&config.TLSKeyPath, "tls-key", "", "Path to the TLS key file")
	// flag.StringVar(&config.TLSCAPath, "tls-ca", "", "Path to the TLS CA certificate file")
	flag.StringVar(&pdAssistantURLs, "pd-assistant-urls", "", "List of PD Assistant URLs (comma-separated). Overrides --pd-assistant-host-prefix and ignores --pd-address auto-discovery if provided")
	flag.StringVar(&certFilePath, "certificate-file", "/app/conf/", "Path to a Certificate YAML file to be used as a template")
	flag.BoolVar(&config.PDAssistantTLSInsecure, "pd-assistant-tls-insecure", false, "Skip TLS verification for PD Assistant instances (not recommended)")
	flag.Parse()

	// Update config
	if err := config.Update(pdAssistantURLs, certFilePath); err != nil {
		glog.Fatalf("Failed to update config: %v", err)
	}

	// Show and exit functions
	if showVersion {
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	// Init k8s client
	kubeClient := k8s.Client{}
	err := kubeClient.Init(kubeconfig)
	if err != nil {
		glog.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Log some useful information
	glog.V(4).Infof("Starting application. Version: %s", Version)
	glog.V(4).Infof("PD Assistant URLs: %v", config.PDAssistantURLs)
	//glog.V(4).Infof("PD Address: %s", config.PDAddress)
	//glog.V(4).Infof("Loaded certificate YAML file %q: name=%s, namespace=%s", certFilePath, config.Certificate.Name, config.Certificate.Namespace)
	//glog.V(4).Infof("TLS Config - Cert: %s, Key: %s, CA: %s", config.TLSCertPath, config.TLSKeyPath, config.TLSCAPath)

	// Let's rock and roll!
	// Watch CliliumNode IPs and update the state
	go srv.IPWatchLoop(config, kubeClient)

	// Watch all pd-assistant IPs and update the certificate if needed
	go srv.FetchIPsAndUpdateCertLoop(config, kubeClient)

	// Start the main web server
	srv.RunMainWebServer(config, listen)
}
