package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/metrics"
	"github.com/impossiblecloud/pd-cert-assistant/internal/tidb"
)

// Constants
var Version string

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

// Main web server
func runMainWebServer(config cfg.AppConfig, listen string) {
	// Setup http router
	router := mux.NewRouter().StrictSlash(true)

	// Routes
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/metrics", handleMetrics(config)).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")

	// Run main http router
	glog.Fatal(http.ListenAndServe(listen, router))
}

func main() {
	var listen string
	var showVersion bool

	// Init config
	config := cfg.AppConfig{}
	config.HTTPRequestTimeout = 5 // seconds

	flag.StringVar(&listen, "listen", ":8765", "Address:port to listen on")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.StringVar(&config.PDAddress, "pd-address", "tidb-cluster-pd:2379", "Address:port of PD server")
	flag.StringVar(&config.TLSCertPath, "tls-cert", "", "Path to the TLS certificate file")
	flag.StringVar(&config.TLSKeyPath, "tls-key", "", "Path to the TLS key file")
	flag.StringVar(&config.TLSCAPath, "tls-ca", "", "Path to the TLS CA certificate file")
	flag.BoolVar(&config.TLSInsecure, "tls-insecure", false, "Skip TLS verification (not recommended)")

	flag.Parse()

	// Show and exit functions
	if showVersion {
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	// Init metric
	config.Metrics = metrics.InitMetrics(Version)

	// Test things
	pdNames, err := tidb.PDGetMemberNames(config)
	if err != nil {
		glog.Fatalf("Failed to get PD member names: %v", err)
	}

	// Log some useful information
	glog.V(4).Infof("Starting application. Version: %s", Version)
	glog.V(4).Infof("PD Address: %s", config.PDAddress)
	glog.V(4).Infof("TLS Config - Cert: %s, Key: %s, CA: %s", config.TLSCertPath, config.TLSKeyPath, config.TLSCAPath)
	glog.V(4).Infof("PD member names: %+v", pdNames)

	runMainWebServer(config, listen)
}
