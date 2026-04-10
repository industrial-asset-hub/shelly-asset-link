/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// FetchComponents calls GET /rpc/Shelly.GetComponents on the given IP address
// and returns the fully parsed ComponentsResponse.
// It is the second-stage enrichment call, performed only for devices already
// validated by GetDeviceInfo.
func FetchComponents(ctx context.Context, client *http.Client, ip string) (ComponentsResponse, error) {
	url := fmt.Sprintf("http://%s/rpc/Shelly.GetComponents", ip)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ComponentsResponse{}, fmt.Errorf("building GetComponents request for %s: %w", ip, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return ComponentsResponse{}, fmt.Errorf("calling GetComponents on %s: %w", ip, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ComponentsResponse{}, fmt.Errorf("reading GetComponents response from %s: %w", ip, err)
	}

	return parseComponentsResponse(body)
}

// rawGetComponentsResponse mirrors the exact JSON envelope returned by the Shelly API.
// Each component is identified by a "key" field; config and status are kept as raw JSON
// until the key is known and the correct target struct can be selected.
type rawGetComponentsResponse struct {
	Components []struct {
		Key    string          `json:"key"`
		Config json.RawMessage `json:"config"`
		Status json.RawMessage `json:"status"`
	} `json:"components"`
}

// parseComponentsResponse unmarshals the raw JSON body into a ComponentsResponse.
// Unknown component keys are silently skipped.
func parseComponentsResponse(body []byte) (ComponentsResponse, error) {
	var raw rawGetComponentsResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return ComponentsResponse{}, fmt.Errorf("parsing GetComponents JSON: %w", err)
	}

	var result ComponentsResponse

	for _, c := range raw.Components {
		switch c.Key {
		case "sys":
			var cfg SysConfig
			var sts SysStatus
			_ = json.Unmarshal(c.Config, &cfg)
			_ = json.Unmarshal(c.Status, &sts)
			result.Sys = &SysComponent{Config: cfg, Status: sts}

		case "eth":
			var cfg EthConfig
			var sts EthStatus
			_ = json.Unmarshal(c.Config, &cfg)
			_ = json.Unmarshal(c.Status, &sts)
			result.Eth = &EthComponent{Config: cfg, Status: sts}

		case "wifi":
			var sts WifiStatus
			_ = json.Unmarshal(c.Status, &sts)
			result.Wifi = &WifiComponent{Status: sts}

		case "bt":
			var cfg BtConfig
			_ = json.Unmarshal(c.Config, &cfg)
			result.Bt = &BtComponent{Config: cfg}

		case "mqtt":
			var cfg MqttConfig
			_ = json.Unmarshal(c.Config, &cfg)
			result.Mqtt = &MqttComponent{Config: cfg}

		default:
			kind, idx, ok := parseFunctionalKey(c.Key)
			if !ok {
				continue
			}
			var cfg ComponentConfig
			_ = json.Unmarshal(c.Config, &cfg)
			result.Functional = append(result.Functional, FunctionalComponent{
				Key:    c.Key,
				Kind:   kind,
				Index:  idx,
				Config: cfg,
			})
		}
	}

	return result, nil
}

// functionalKinds contains the component key prefixes that map to IAH functional_parts.
var functionalKinds = map[string]bool{
	"switch": true,
	"cover":  true,
	"light":  true,
}

// parseFunctionalKey parses a component key like "switch:0" into its kind ("switch"),
// index (0), and reports whether the key belongs to a functional component.
func parseFunctionalKey(key string) (kind string, index int, ok bool) {
	colon := strings.IndexByte(key, ':')
	if colon < 0 {
		return "", 0, false
	}
	k := key[:colon]
	if !functionalKinds[k] {
		return "", 0, false
	}
	idx, err := strconv.Atoi(key[colon+1:])
	if err != nil {
		return "", 0, false
	}
	return k, idx, true
}
