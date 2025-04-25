package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/k8s"
	"github.com/impossiblecloud/pd-cert-assistant/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// State holds the state of the application
type State struct {
	// IPAddresses holds the list of Cilium node IP addresses
	IPAddresses []string
	// AllIPAddresses holds the list of all IP addresses from add pd-advisor instances
	AllIPAddresses []string
	// Metrics contains the application's metrics.
	Metrics metrics.AppMetrics
}

// Prometheus metrics handler
func (s *State) handleMetrics(config cfg.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		glog.V(10).Info("Got HTTP request for /metrics")

		promhttp.HandlerFor(prometheus.Gatherer(s.Metrics.Registry), promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}

// Root handler
func rootHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(10).Info("Got HTTP request for /")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Up and running")
}

// Health handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(10).Info("Got HTTP request for /health")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Health is OK")
}

// Auth decorator for all endpoints that require authentication
func authHandler(endpoint http.HandlerFunc, cfg cfg.AppConfig) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Extract the Authorization header
		authHeader := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authHeader) != 2 || authHeader[0] != "Bearer" {
			glog.Warning("Missing or invalid Authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error": "Unauthorized"}`)
			return
		}

		// Validate the token (replace "your-secret-token" with your actual token)
		if authHeader[1] != cfg.BearerToken {
			glog.Warning("Invalid bearer token")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error": "Unauthorized"}`)
			return
		}
		// Call the original endpoint handler
		endpoint(w, r)
	})
}

// IPWatchLoop continuously fetches CiliumNode IPs and updates the state
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

// AllIPsFetchLoop continuously fetches IPs from all pd-assistant instances and updates the state
func (s *State) AllIPsFetchLoop(conf cfg.AppConfig) {
	for {
		// TODO: implement the logic to fetch IPs from all pd-assistant instances
		s.AllIPAddresses = s.IPAddresses
		glog.V(6).Infof("Updated All IPs to: %+v", s.AllIPAddresses)
		// Sleep for a while before the next iteration
		time.Sleep(time.Duration(conf.PDAssistantPollInterval) * time.Second)
	}
}

// UpdateCertLoop continuously checks and updates the certificate based on the IPs stored in the state.
func (s *State) UpdateCertLoop(conf cfg.AppConfig, kc k8s.Client) {
	for {
		glog.V(6).Infof("Checking for certificate updates...")
		err := kc.UpdateCertificate(s.AllIPAddresses)
		if err != nil {
			s.Metrics.CertUpdateErrors.WithLabelValues().Inc()
			glog.Errorf("Failed to update certificate: %v", err)
		}

		// Sleep for a while before the next iteration
		time.Sleep(time.Duration(conf.KubernetesPollInterval) * time.Second)
	}
}

// GetIPs handler with bearer token authentication
func (s *State) GetIPs(w http.ResponseWriter, r *http.Request) {
	glog.V(10).Info("Got HTTP request for /api/ips")

	// Marshal the IP addresses to JSON
	jsonResponse, err := json.Marshal(s.IPAddresses)
	if err != nil {
		glog.Errorf("Failed to marshal IP addresses: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to encode IP addresses"}`)
		return
	}

	// Respond with the IP addresses
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
	router.HandleFunc("/metrics", s.handleMetrics(config)).Methods("GET")
	router.HandleFunc("/api/ips", authHandler(s.GetIPs, config)).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")

	// Run main http router
	glog.Fatal(http.ListenAndServe(listen, router))
}
