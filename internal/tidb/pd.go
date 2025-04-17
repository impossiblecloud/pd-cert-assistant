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
	resp, err := utils.MakeHTTPRequest(pdAddress, conf.TLSCertPath, conf.TLSKeyPath, conf.TLSCAPath, conf.TLSInsecure, conf.HTTPRequestTimeout)
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
