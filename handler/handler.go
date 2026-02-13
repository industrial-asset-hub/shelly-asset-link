/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/industrial-asset-hub/asset-link-sdk/v3/config"
	generated "github.com/industrial-asset-hub/asset-link-sdk/v3/generated/iah-discovery"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/model"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/publish"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"shelly-asset-link/internal/shellycfg"
	"shelly-asset-link/scanner"
)

type AssetLinkImplementation struct {
	discoveryLock sync.Mutex
}

func (m *AssetLinkImplementation) Discover(discoveryConfig config.DiscoveryConfig, devicePublisher publish.DevicePublisher) error {
	log.Info().Msg("Handle Discovery Request")

	// Concurrency guard
	if m.discoveryLock.TryLock() {
		defer m.discoveryLock.Unlock()
	} else {
		return status.Errorf(codes.ResourceExhausted, "Another discovery job is already running")
	}

	// --- shelly-AL specific: read IPRange from request; robust via protojson -> extract to json ---
	var ipRange string
	if req := discoveryConfig.GetDiscoveryRequest(); req != nil {
		for _, f := range req.GetFilters() {
			k := strings.TrimSpace(f.GetKey())
			if strings.EqualFold(k, "IPRange") {
				if s, ok := variantAsStringJSON(f.GetValue()); ok {
					ipRange = s
				} else {
					log.Warn().Msg("IPRange filter present but empty/unsupported variant")
				}
			}
		}
	}
	log.Info().Msgf("Discovery Config liefert: ipRange=%q", ipRange)

	// --- shelly-AL specific: only take ipRange from protobuff, rest via fileCfg/Defaults in BuildFromInputs ---
	in := shellycfg.Inputs{
		IpRange: ipRange,
		// Cidr/Subnet/StartIP/EndIP/TimeoutMs/MaxParallel -> via merge/Defaults
		// handover of other filters and options subject to later changes
	}

	// --- shelly-AL specific: read config.json as fallback/defaults
	path := os.Getenv("AL_CONFIG_PATH")
	if path == "" {
		path = "/config/shelly_al_config.json"
	}
	fileCfg := shellycfg.LoadFileConfig(path)

	// --- build effective ScanConfig (final deduction and sanity happen in in merge.go)
	cfg := shellycfg.BuildFromInputs(in, fileCfg)
	log.Info().Msgf("Effective ScanConfig aus merge.go: {Subnet:%s StartIP:%d EndIP:%d Timeout:%s MaxParallel:%d}",
		cfg.Subnet, cfg.StartIP, cfg.EndIP, cfg.Timeout, cfg.MaxParallel)

	// --- shelly-AL: make sure timeout is no bogus
	scanTimeout := cfg.Timeout
	if scanTimeout <= 0 {
		scanTimeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	// --- shelly-AL: finally fire shellyscanner.go with context so it can be stopped and configuration
	results, err := scanner.RunScan(ctx, cfg)
	if err != nil {
		log.Error().Err(err).Msg("Scan failed")
		return status.Errorf(codes.Unavailable, "Scan failed: %v", err)
	}

	// --- shelly-AL: publish results
	for _, r := range results {
		name := "Shelly Device"
		if r.DeviceInfo.Name != nil && *r.DeviceInfo.Name != "" {
			name = *r.DeviceInfo.Name
		}

		deviceInfo := model.NewDevice("EthernetDevice", name)
		deviceInfo.AddNameplate(
			"Shelly",
			"",
			r.DeviceInfo.Model,
			fmt.Sprintf("Shelly Appliance=%s Gen%d", r.DeviceInfo.App, r.DeviceInfo.Gen),
			r.DeviceInfo.Gen,
			r.DeviceInfo.ID,
		)
		deviceInfo.AddSoftware("firmware", r.DeviceInfo.FwID, true) // switched to lower case for SSG compatibility

		nicID := deviceInfo.AddNic("wifi", r.DeviceInfo.Mac)
		if wifi, ok := r.Status["wifi"].(map[string]interface{}); ok {
			if ip, ok := wifi["sta_ip"].(string); ok && ip != "" {
				deviceInfo.AddIPv4(nicID, ip, "255.255.255.0", "")
			} else {
				deviceInfo.AddIPv4(nicID, r.IPAddress, "255.255.255.0", "")
			}
		} else {
			deviceInfo.AddIPv4(nicID, r.IPAddress, "255.255.255.0", "")
		}

		discoveredDevice := deviceInfo.ConvertToDiscoveredDevice()
		if err := devicePublisher.PublishDevice(discoveredDevice); err != nil {
			return err
		}
	}

	return nil
}

// GetSupportedOptions hält die API-Kompatibilität; der Handler nutzt aktuell nur IPRange
func (m *AssetLinkImplementation) GetSupportedOptions() []*generated.SupportedOption {
	return []*generated.SupportedOption{
		{Key: "timeout", Datatype: generated.VariantType_VT_STRING},
		{Key: "timeoutMs", Datatype: generated.VariantType_VT_STRING},
		{Key: "maxParallel", Datatype: generated.VariantType_VT_STRING},
	}
}

// GetSupportedFilters deklariert IPRange als String (UI-Vertrag)
func (m *AssetLinkImplementation) GetSupportedFilters() []*generated.SupportedFilter {
	return []*generated.SupportedFilter{
		{Key: "IPRange", Datatype: generated.VariantType_VT_STRING},
	}
}

// --- Helper: extract variant robustly as string ---
// strategy: variant -> protojson.Marshal -> generic JSON -> take the first string that makes sense
// supports base64-coded bytes/ rawData as well
func variantAsStringJSON(v *generated.Variant) (string, bool) {
	if v == nil {
		return "", false
	}

	// to JSON
	b, err := protojson.MarshalOptions{EmitUnpopulated: false}.Marshal(v)
	if err != nil || len(b) == 0 {
		return "", false
	}

	// parse
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return "", false
	}

	// find first meaningful value
	if s, ok := jsonAnyToString(m); ok && strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s), true
	}
	return "", false
}

// jsonAnyToString looks in generic JSON-structs for meaningful string
// order: string (base64->text) > number > bool > rekursively in maps/lists
func jsonAnyToString(x any) (string, bool) {
	switch v := x.(type) {
	case string:
		if s2, ok := tryBase64ToText(v); ok {
			return s2, true
		}
		return v, strings.TrimSpace(v) != ""
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%g", v)), true
	case bool:
		if v {
			return "true", true
		}
		return "false", true
	case map[string]any:
		// 1) prioritise known keys
		for _, key := range []string{
			"stringValue", "string_value", "string",
			"rawData", "raw_data", "bytes",
			"intValue", "int_value", "number", "num",
			"boolValue", "bool_value", "bool",
			"doubleValue", "double_value", "double", "float",
			"value",
		} {
			if val, ok := v[key]; ok {
				if s, ok := jsonAnyToString(val); ok && strings.TrimSpace(s) != "" {
					return s, true
				}
			}
		}
		// 2) else: first meaningful field
		for _, val := range v {
			if s, ok := jsonAnyToString(val); ok && strings.TrimSpace(s) != "" {
				return s, true
			}
		}
		return "", false
	case []any:
		// list: take first element
		for _, elem := range v {
			if s, ok := jsonAnyToString(elem); ok && strings.TrimSpace(s) != "" {
				return s, true
			}
		}
		return "", false
	default:
		return "", false
	}
}

// tryBase64ToText: decode base64-string to text
// response: (text, true) if decoding makes sense
func tryBase64ToText(s string) (string, bool) {
	// short heuristics-check: if s too short, skip decoding
	if len(s) < 8 {
		return "", false
	}
	// try std base64
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(data) == 0 {
		return "", false
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", false
	}
	// simple printability check
	printable := 0
	for _, r := range text {
		if r >= 32 && r <= 126 {
			printable++
		}
	}
	if printable == 0 {
		return "", false
	}
	return text, true
}
