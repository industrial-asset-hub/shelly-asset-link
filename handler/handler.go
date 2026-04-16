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
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/industrial-asset-hub/asset-link-sdk/v3/config"
	generated "github.com/industrial-asset-hub/asset-link-sdk/v3/generated/iah-discovery"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/publish"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"shelly-asset-link/internal/shellycfg"
	"shelly-asset-link/mapping"
	"shelly-asset-link/scanner"
	"shelly-asset-link/shelly"
)

// AssetLinkImplementation implements the IAH Discovery interface.
type AssetLinkImplementation struct {
	discoveryLock sync.Mutex
}

// Discover handles a discovery request from the Industrial Asset Hub.
//
// It orchestrates the two-stage pipeline:
//  1. scanner.RunScan     — validates IPs via GetDeviceInfo (first-stage filter)
//  2. shelly.FetchComponents — enriches each confirmed device via GetComponents
//  3. mapping.Map         — transforms Shelly data into the IAH base schema
//  4. DevicePublisher     — streams each device immediately to the IAH
func (m *AssetLinkImplementation) Discover(discoveryConfig config.DiscoveryConfig, devicePublisher publish.DevicePublisher) error {
	log.Info().Msg("Handle Discovery Request")

	// Concurrency guard: only one discovery at a time
	if m.discoveryLock.TryLock() {
		defer m.discoveryLock.Unlock()
	} else {
		return status.Errorf(codes.ResourceExhausted, "Another discovery job is already running")
	}

	// Extract filters and options from the Discovery request
	var ipRange string
	var scanTimeoutMs, httpTimeoutMs *int
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
		for _, o := range req.GetOptions() {
			k := strings.TrimSpace(o.GetKey())
			s, ok := variantAsStringJSON(o.GetValue())
			switch {
			case strings.EqualFold(k, "scanTimeoutMs"):
				if ok {
					if v, err := strconv.Atoi(s); err == nil && v > 0 {
						scanTimeoutMs = &v
					} else {
						log.Warn().Msgf("scanTimeoutMs option invalid value %q — ignored", s)
					}
				}
			case strings.EqualFold(k, "httpTimeoutMs"):
				if ok {
					if v, err := strconv.Atoi(s); err == nil && v > 0 {
						httpTimeoutMs = &v
					} else {
						log.Warn().Msgf("httpTimeoutMs option invalid value %q — ignored", s)
					}
				}
			}
		}
	}
	log.Info().Msgf("Discovery Config: ipRange=%q", ipRange)

	in := shellycfg.Inputs{
		IpRange:       ipRange,
		ScanTimeoutMs: scanTimeoutMs,
		HttpTimeoutMs: httpTimeoutMs,
	}

	// Load file config as the baseline configuration source
	path := os.Getenv("AL_CONFIG_PATH")
	if path == "" {
		path = "/config/shelly_al_config.json"
	}
	fileCfg := shellycfg.LoadFileConfig(path)

	// Build effective ScanConfig; fails when no valid IP range can be determined
	cfg, err := shellycfg.BuildFromInputs(in, fileCfg)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid scan configuration: %v", err)
	}
	log.Info().Msgf("Effective ScanConfig: {Subnet:%s StartIP:%d EndIP:%d ScanTimeout:%s HTTPTimeout:%s}",
		cfg.Subnet, cfg.StartIP, cfg.EndIP, cfg.ScanTimeout, cfg.HTTPTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ScanTimeout)
	defer cancel()

	// Stage 1: scan — GetDeviceInfo per IP, only confirmed Shelly devices pass through
	deviceCh, errCh := scanner.RunScan(ctx, cfg)

	// Shared HTTP client for stage-2 enrichment calls (short per-request timeout)
	client := &http.Client{Timeout: cfg.HTTPTimeout}

	for dev := range deviceCh {
		// Stage 2: enrich — GetComponents for each validated device
		comps, fetchErr := shelly.FetchComponents(ctx, client, dev.IPAddress)
		if fetchErr != nil {
			log.Warn().Err(fetchErr).Msgf("GetComponents failed for %s — skipping device", dev.IPAddress)
			continue
		}

		// Stage 3: transform — pure mapping from Shelly model to IAH base schema
		device := mapping.Map(dev.DeviceInfo, comps, dev.IPAddress)

		// Stage 4: publish — stream immediately to IAH, no batch accumulation
		if pubErr := devicePublisher.PublishDevice(device.ConvertToDiscoveredDevice()); pubErr != nil {
			return pubErr
		}
	}

	// Check for a fatal error reported by the scanner goroutine after the channel is drained
	if scanErr := <-errCh; scanErr != nil {
		log.Error().Err(scanErr).Msg("Scanner reported a fatal error")
	}

	return nil
}

// GetSupportedFilters declares the accepted discovery filter:
//   - IPRange: subnet scan range (e.g. "10.0.20.30-101")
func (m *AssetLinkImplementation) GetSupportedFilters() []*generated.SupportedFilter {
	return []*generated.SupportedFilter{
		{Key: "IPRange", Datatype: generated.VariantType_VT_STRING},
	}
}

// GetSupportedOptions declares the accepted discovery options:
//   - scanTimeoutMs: maximum total scan duration in milliseconds
//   - httpTimeoutMs: per-request HTTP timeout for individual Shelly API calls in milliseconds
func (m *AssetLinkImplementation) GetSupportedOptions() []*generated.SupportedOption {
	return []*generated.SupportedOption{
		{Key: "scanTimeoutMs", Datatype: generated.VariantType_VT_STRING},
		{Key: "httpTimeoutMs", Datatype: generated.VariantType_VT_STRING},
	}
}

// variantAsStringJSON extracts a string value from a Variant via protojson marshalling.
// It also handles base64-encoded byte values.
func variantAsStringJSON(v *generated.Variant) (string, bool) {
	if v == nil {
		return "", false
	}

	b, err := protojson.MarshalOptions{EmitUnpopulated: false}.Marshal(v)
	if err != nil || len(b) == 0 {
		return "", false
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return "", false
	}

	if s, ok := jsonAnyToString(m); ok && strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s), true
	}
	return "", false
}

// jsonAnyToString recursively searches generic JSON structures for a meaningful string value.
// Priority: string (base64→text) > number > bool > map keys > list elements.
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
		for _, val := range v {
			if s, ok := jsonAnyToString(val); ok && strings.TrimSpace(s) != "" {
				return s, true
			}
		}
		return "", false
	case []any:
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

// tryBase64ToText attempts to decode a base64 string to printable text.
func tryBase64ToText(s string) (string, bool) {
	if len(s) < 8 {
		return "", false
	}
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(data) == 0 {
		return "", false
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", false
	}
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
