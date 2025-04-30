package tidb

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/glog"

	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/utils"
)

func encodePDDiscoveryPath(conf cfg.PDDiscoveryConfig) string {
	domain := fmt.Sprintf("%s-pd-0.%s-pd-peer.%s.svc:2380", conf.TiDBCLusterName, conf.TiDBCLusterName, conf.TiDBCLusterNameSpace)
	encoded := base64.StdEncoding.EncodeToString([]byte(domain))
	return strings.ReplaceAll(encoded, "\n", "")
}

// PDGetMemberNames fetches a list of members from a PD server and returns their names.
func PDGetMemberNames(conf cfg.PDConfig) ([]string, error) {
	pdScheme := "http://"
	if conf.TLSConfig.CAPath != "" {
		pdScheme = "https://"
	}
	pdAddress := pdScheme + conf.Address + "/pd/api/v1/members"
	resp, err := utils.MakeHTTPRequest(pdAddress, conf.TLSConfig.CertPath, conf.TLSConfig.KeyPath, conf.TLSConfig.CAPath, conf.TLSConfig.Insecure, conf.HTTPRequestTimeout, "")
	// Check if the request was successful
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to %q: %v", pdAddress, err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	// Parse the JSON response
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	// Extract ".members[].name" values
	members, ok := data["members"].([]interface{})
	if !ok {
		return nil, errors.New("invalid JSON structure: 'members' field is missing or not an array")
	}

	var names []string
	for _, member := range members {
		memberMap, ok := member.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := memberMap["name"].(string); ok {
			names = append(names, name)
		}
	}

	return names, nil
}

// PDDiscoveryGetMemberNames fetches a list of members from a PD discovery service and returns their names.
func PDDiscoveryGetMemberNames(conf cfg.PDDiscoveryConfig) ([]string, error) {
	pdDiscoveryPath := encodePDDiscoveryPath(conf)
	pdDiscoveryURL := fmt.Sprintf("%s/new/%s", conf.URL, pdDiscoveryPath)
	resp, err := utils.MakeHTTPRequest(pdDiscoveryURL, conf.TLSConfig.CertPath, conf.TLSConfig.KeyPath, conf.TLSConfig.CAPath, conf.TLSConfig.Insecure, conf.HTTPRequestTimeout, "")
	// Check if the request was successful
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to %q: %v", pdDiscoveryURL, err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status from %q: %s", pdDiscoveryURL, resp.Status)
	}

	// Convert response body to string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	bodyString := string(bodyBytes)
	glog.V(8).Infof("PD Discovery Response body: %s", bodyString)

	hosts := []string{}
	for _, url := range utils.FindUniqueURLs(bodyString) {
		glog.V(8).Infof("Found PD URL: %s", url)
		host := utils.GetHostFromURL(url)
		if host == "" {
			continue
		}
		if !utils.Contains(hosts, host) {
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}

// GetUniqueDomains extracts unique domains from a list of hosts.
func GetUniqueDomains(hosts []string) []string {
	var result []string
	for _, host := range hosts {
		domain := utils.GetDomainFromHost(host)
		if domain == "" {
			continue
		}
		if !utils.Contains(result, domain) {
			result = append(result, domain)
		}
	}
	return result
}

// BuildPDAssistantHostnames generates PD Assistant hostnames based on the provided configuration and domains.
func BuildPDAssistantHostnames(conf cfg.AppConfig, hosts []string) []string {
	var pdAssistantHosts []string
	domains := GetUniqueDomains(hosts)
	for _, domain := range domains {
		pdAssistantHosts = append(pdAssistantHosts, conf.PDAssistantHostPrefix+"."+domain)
	}
	return pdAssistantHosts
}

// GetPDAssistantURLs retrieves PD Assistant hostnames based on the PD member names and generates a list of URLs.
func GetPDAssistantURLs(conf cfg.AppConfig) ([]string, error) {
	pdNames, err := PDDiscoveryGetMemberNames(conf.PDDiscoveryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get PD member names: %v", err)
	}

	pdAssistantHosts := BuildPDAssistantHostnames(conf, pdNames)
	if len(pdAssistantHosts) == 0 {
		return nil, errors.New("no PD Assistant hostnames found")
	}

	pdAssistanURLs := []string{}
	for _, host := range pdAssistantHosts {
		pdAssistanURLs = append(pdAssistanURLs, fmt.Sprintf("%s://%s:%s", conf.PDAssistantScheme, host, conf.PDAssistantPort))
	}

	return pdAssistanURLs, nil
}

// GetPDDAssistantURLs retrieves PD Assistant hostnames based on the PD member names from discovery service and generates a list of URLs.
func GetPDDiscoveryAssistantURLs(conf cfg.AppConfig) ([]string, error) {
	pdNames, err := PDDiscoveryGetMemberNames(conf.PDDiscoveryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get PD member names: %v", err)
	}

	pdAssistantHosts := BuildPDAssistantHostnames(conf, pdNames)
	if len(pdAssistantHosts) == 0 {
		return nil, errors.New("no PD Assistant hostnames found")
	}

	pdAssistanURLs := []string{}
	for _, host := range pdAssistantHosts {
		pdAssistanURLs = append(pdAssistanURLs, fmt.Sprintf("%s://%s:%s", conf.PDAssistantScheme, host, conf.PDAssistantPort))
	}

	return pdAssistanURLs, nil
}
