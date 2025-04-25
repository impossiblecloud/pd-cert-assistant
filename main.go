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
	"github.com/impossiblecloud/pd-cert-assistant/internal/tidb"
)

// Constants
var Version string

func main() {
	var listen, kubeconfig string
	var showVersion bool

	if Version == "" {
		Version = "unknown"
	}

	// Init config
	config := cfg.AppConfig{}
	config.HTTPRequestTimeout = 5 // seconds
	config.BearerToken = os.Getenv("BEARER_TOKEN")
	if config.BearerToken == "" {
		glog.Fatal("BEARER_TOKEN environment variable is not set")
	}

	// Init state
	srv := server.State{}
	// Init metric
	srv.Metrics = metrics.InitMetrics(Version)

	flag.StringVar(&listen, "listen", ":8765", "Address:port to listen on")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.StringVar(&config.PDAddress, "pd-address", "tidb-cluster-pd:2379", "Address:port of PD server")
	flag.StringVar(&config.TLSCertPath, "tls-cert", "", "Path to the TLS certificate file")
	flag.StringVar(&config.TLSKeyPath, "tls-key", "", "Path to the TLS key file")
	flag.StringVar(&config.TLSCAPath, "tls-ca", "", "Path to the TLS CA certificate file")
	flag.BoolVar(&config.TLSInsecure, "tls-insecure", false, "Skip TLS verification (not recommended)")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file (optional)")
	flag.IntVar(&config.KubernetesPollInterval, "k8s-poll-interval", 180, "Interval for polling Kubernetes in seconds")
	flag.IntVar(&config.PDAssistantPollInterval, "pd-assistant-poll-interval", 300, "Interval for polling all pd-assistants in seconds")
	flag.IntVar(&config.CertUpdateInterval, "cert-update-interval", 300, "Interval for updating PD certificate in seconds")
	flag.StringVar(&config.PDAssistantHostPrefix, "pd-assistant-host-prefix", "pd-assistant", "Host prefix for PD Assistant instances")
	flag.StringVar(&config.CertificateName, "cert-name", "pd-assistant", "Name of the certificate to be used")

	flag.Parse()

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
	glog.V(4).Infof("PD Address: %s", config.PDAddress)
	glog.V(4).Infof("TLS Config - Cert: %s, Key: %s, CA: %s", config.TLSCertPath, config.TLSKeyPath, config.TLSCAPath)

	// Test things
	pdNames, err := tidb.PDGetMemberNames(config)
	if err != nil {
		glog.Fatalf("Failed to get PD member names: %v", err)
	}
	pdAssistants := tidb.BuildPDAssistantHostnames(config, pdNames)

	glog.V(4).Infof("PD member names: %+v", pdNames)
	glog.V(4).Infof("PD assistants: %+v", pdAssistants)

	// Let's rock and roll!
	// Watch CliliumNode IPs and update the state
	go srv.IPWatchLoop(config, kubeClient)

	// Watch all pd-assistant IPs and update the state
	go srv.AllIPsFetchLoop(config)

	// Update the certificate based on the IPs
	go srv.UpdateCertLoop(config, kubeClient)

	// Start the main web server
	srv.RunMainWebServer(config, listen)
}
