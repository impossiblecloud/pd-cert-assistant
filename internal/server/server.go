package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/impossiblecloud/pd-cert-assistant/internal/api"
	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/k8s"
	"github.com/impossiblecloud/pd-cert-assistant/internal/metrics"
	"github.com/impossiblecloud/pd-cert-assistant/internal/tidb"
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

func (s *State) getAllIPAddresses(conf cfg.AppConfig, pdaAddresses []string) ([]string, error) {
	allIPAddresses := []string{}
	for _, pdaAddress := range pdaAddresses {
		glog.V(6).Infof("Fetching IPs from pd-assistant: %s", pdaAddress)
		ips, err := api.GetIPs(conf, pdaAddress)
		if err != nil {
			s.Metrics.PDAssistantFetchErrors.WithLabelValues(pdaAddress).Inc()
			return nil, fmt.Errorf("failed to fetch IPs from pd-assistant %s: %v", pdaAddress, err)

		}
		if len(ips) == 0 {
			s.Metrics.PDAssistantFetchErrors.WithLabelValues(pdaAddress).Inc()
			return nil, fmt.Errorf("no IPs found in pd-assistant %s", pdaAddress)
		}

		// Update the state with the fetched IPs
		allIPAddresses = append(allIPAddresses, ips...)
		glog.V(6).Infof("Fetched IPs from pd-assistant %s: %+v", pdaAddress, ips)
	}
	return allIPAddresses, nil
}

// IPWatchLoop continuously fetches CiliumNode IPs and updates the state
func (s *State) IPWatchLoop(conf cfg.AppConfig, kc k8s.Client) {
	for {
		ciliumNodeIPs, err := kc.GetCiliumNodes()
		if err != nil {
			glog.Errorf("Failed to fetch CiliumNodes: %v", err)
		} else {
			s.IPAddresses = ciliumNodeIPs
			s.Metrics.LocalIPs.WithLabelValues().Set(float64(len(ciliumNodeIPs)))
			glog.V(6).Infof("Updated State IPs to: %+v", ciliumNodeIPs)
		}

		// Sleep for a while before the next iteration
		time.Sleep(time.Duration(conf.KubernetesPollInterval) * time.Second)
	}
}

// AllIPsFetchLoop continuously fetches IPs from all pd-assistant instances and updates the state
func (s *State) FetchIPsAndUpdateCertLoop(conf cfg.AppConfig, kc k8s.Client) {
	for {
		// Sleep before iteration
		time.Sleep(time.Duration(conf.PDAssistantPollInterval) * time.Second)

		// Do stuff
		pdaAddresses := conf.PDAssistantAddresses
		if len(conf.PDAssistantAddresses) == 0 {
			// If no pd-assistant addresses are provided, fetch them from the PD server
			var err error
			pdaAddresses, err = tidb.GetPDAssistantURLs(conf)
			if err != nil {
				glog.Errorf("Failed to fetch PD Assistant URLs: %s", err.Error())
				// It's unsafe to continue if we can't fetch IPs, so we log the error and skip this iteration
				continue
			}
		}
		allIPAddresses, err := s.getAllIPAddresses(conf, pdaAddresses)
		if err != nil {
			glog.Errorf("Failed to fetch IPs from pd-assistants: %v", err)
			// It's unsafe to continue if we can't fetch IPs, so we log the error and skip this iteration
			continue
		}

		// Failsafe check for empty IPs, we should never have empty IPs
		if len(allIPAddresses) == 0 {
			glog.Errorf("No IPs found in pd-assistants")
			continue
		}

		// Atomic update of AllIPAddresses in the state, only if all IPs are fetched successfully
		s.AllIPAddresses = allIPAddresses
		s.Metrics.AllIPs.WithLabelValues().Set(float64(len(allIPAddresses)))
		glog.V(4).Infof("Updated All IPs to: %+v", s.AllIPAddresses)
		glog.V(6).Infof("Checking for certificate updates, certificate name: %s", conf.CertificateName)

		// Update the certificate with the new IPs if needed
		err = kc.UpdateCertificate(conf, allIPAddresses)
		if err != nil {
			s.Metrics.CertUpdateErrors.WithLabelValues().Inc()
			glog.Errorf("Failed to update certificate: %v", err)
		}
	}
}

// GetIPs handler with bearer token authentication
func (s *State) GetIPs(w http.ResponseWriter, r *http.Request) {
	glog.V(10).Infof("Got HTTP request for %s", api.ApiIPsPath)

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
	router.HandleFunc(api.ApiIPsPath, authHandler(s.GetIPs, config)).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")

	// Run main http router
	glog.Fatal(http.ListenAndServe(listen, router))
}
