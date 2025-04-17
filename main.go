package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/k8s"
	"github.com/impossiblecloud/pd-cert-assistant/internal/metrics"
	"github.com/impossiblecloud/pd-cert-assistant/internal/tidb"
)

// Constants
var Version string

// State holds the state of the application
type State struct {
	IPAddresses []string
}

// Prometheus metrics handler
func handleMetrics(config cfg.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		glog.V(10).Info("Got HTTP request for /metrics")

		promhttp.HandlerFor(prometheus.Gatherer(config.Metrics.Registry), promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}

// Root handler
func rootHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(10).Info("Got HTTP request for /")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Up and running. Version: %s", Version)
}

// Health handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(10).Info("Got HTTP request for /health")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Health is OK")
}

func (s *State) IPWatchLoop(conf cfg.AppConfig, kc k8s.Client) {
	for {
		ciliumNodeIPs, err := kc.GetCiliumNodes()
		if err != nil {
			glog.Errorf("Failed to fetch CiliumNodes: %v", err)
		} else {
			s.IPAddresses = ciliumNodeIPs
			glog.V(6).Infof("Updated State IPs to: %+v", ciliumNodeIPs)
		}

		// Sleep for a while before the next iteration
		time.Sleep(time.Duration(conf.KubernetesPollInterval) * time.Second)
	}
}

// GetIPs handler
func (s *State) GetIPs(w http.ResponseWriter, r *http.Request) {
	glog.V(10).Info("Got HTTP request for /health")

	jsonResponse, err := json.Marshal(s.IPAddresses)
	if err != nil {
		glog.Errorf("Failed to marshal IP addresses: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to encode IP addresses"}`)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// Main web server
func (s *State) RunMainWebServer(config cfg.AppConfig, listen string) {
	// Setup http router
	router := mux.NewRouter().StrictSlash(true)

	// Routes
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/metrics", handleMetrics(config)).Methods("GET")
	router.HandleFunc("/api/ips", s.GetIPs).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")

	// Run main http router
	glog.Fatal(http.ListenAndServe(listen, router))
}

func main() {
	var listen, kubeconfig string
	var showVersion bool

	if Version == "" {
		Version = "unknown"
	}

	// Init config and state
	config := cfg.AppConfig{}
	config.HTTPRequestTimeout = 5 // seconds
	state := State{}

	flag.StringVar(&listen, "listen", ":8765", "Address:port to listen on")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.StringVar(&config.PDAddress, "pd-address", "tidb-cluster-pd:2379", "Address:port of PD server")
	flag.StringVar(&config.TLSCertPath, "tls-cert", "", "Path to the TLS certificate file")
	flag.StringVar(&config.TLSKeyPath, "tls-key", "", "Path to the TLS key file")
	flag.StringVar(&config.TLSCAPath, "tls-ca", "", "Path to the TLS CA certificate file")
	flag.BoolVar(&config.TLSInsecure, "tls-insecure", false, "Skip TLS verification (not recommended)")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file (optional)")
	flag.IntVar(&config.KubernetesPollInterval, "k8s-poll-interval", 180, "Interval for polling Kubernetes in seconds")

	flag.Parse()

	// Show and exit functions
	if showVersion {
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	// Init metric
	config.Metrics = metrics.InitMetrics(Version)

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
	domains := tidb.GetUniqueDomains(pdNames)

	glog.V(4).Infof("PD member names: %+v", pdNames)
	glog.V(4).Infof("Unique TiDB cluster domains: %+v", domains)

	go state.IPWatchLoop(config, kubeClient)
	state.RunMainWebServer(config, listen)
}
