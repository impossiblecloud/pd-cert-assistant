package api

import (
	"fmt"

	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/utils"
)

const (
	ApiIPsPath    = "/api/v1/ips"
	ApiAllIPsPath = "/api/v1/allips"
)

// For now we don't really have any API, just parsing JSON response with []string data in it.
func getIPs(conf cfg.AppConfig, pdaAddress, path string) ([]string, error) {
	fullAddress := pdaAddress + path
	resp, err := utils.MakeHTTPRequest(fullAddress, "", "", "", conf.PDAssistantTLSInsecure, conf.HTTPRequestTimeout, conf.BearerToken)
	// Check if the request was successful
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTPS request to %s: %s", pdaAddress, err.Error())
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received non-OK HTTP status from %s: %s", pdaAddress, resp.Status)
	}

	// Parse the JSON response
	var ips []string
	if err := utils.ParseJSONResponse(resp.Body, &ips); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from %s: %s", pdaAddress, err.Error())
	}

	return ips, nil
}

// GetIPs fetches local IP addresses from the PD Assistant instances.
func GetLocalIPs(conf cfg.AppConfig, pdaAddress string) ([]string, error) {
	return getIPs(conf, pdaAddress, ApiIPsPath)
}

// GetIAllPs fetches all IP addresses from the PD Assistant instances.
func GetAllIPs(conf cfg.AppConfig, pdaAddress string) ([]string, error) {
	return getIPs(conf, pdaAddress, ApiAllIPsPath)
}
