package tidb

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/utils"
)

// PDGetMemberNames fetches a list of members from a PD server and returns their names.
func PDGetMemberNames(conf cfg.AppConfig) ([]string, error) {
	pdScheme := "http://"
	if conf.TLSCAPath != "" {
		pdScheme = "https://"
	}
	pdAddress := pdScheme + conf.PDAddress + "/pd/api/v1/members"
	resp, err := utils.MakeHTTPRequest(pdAddress, conf.TLSCertPath, conf.TLSKeyPath, conf.TLSCAPath, conf.TLSInsecure, conf.HTTPRequestTimeout, "")
	// Check if the request was successful
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTPS request: %v", err)
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
	pdNames, err := PDGetMemberNames(conf)
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
