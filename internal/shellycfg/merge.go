/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package shellycfg

import (
	"fmt"
	"strings"
	"time"

	"shelly-asset-link/scanner"
)

// BuildFromInputs merges runtime inputs (Discovery filters/options) with the static
// file config (fileCfg) and returns the effective ScanConfig for the scanner.
//
// Priority order (highest to lowest):
//  1. Dynamic filter values from the Discovery request (IPRange, scanTimeoutMs, httpTimeoutMs)
//  2. File config (shelly_al_config.json)
//  3. Hardcoded defaults (ScanTimeout: 300s, HTTPTimeout: 2s)
//
// Returns an error when no valid IP range can be determined from the available sources.
func BuildFromInputs(in Inputs, fileCfg *FileScanConfig) (scanner.ScanConfig, error) {
	var subnet string
	var start, end, scanTimeoutMs, httpTimeoutMs int

	// Layer 1: seed from file config when available
	if fileCfg != nil {
		subnet = fileCfg.Subnet
		start = fileCfg.StartIP
		end = fileCfg.EndIP
		scanTimeoutMs = fileCfg.ScanTimeoutMs
		httpTimeoutMs = fileCfg.HttpTimeoutMs
	}

	// Layer 2: Dynamic filter values from the Discovery request take precedence
	if s := strings.TrimSpace(in.IpRange); s != "" {
		if sn, st, en, ok := ParseIPRange(s); ok {
			subnet, start, end = sn, st, en
		}
		// On parse error the file config values stay active
	} else {
		// Explicit subnet / start / end fields from Inputs
		if strings.TrimSpace(in.Subnet) != "" {
			subnet = strings.TrimSuffix(strings.TrimSpace(in.Subnet), ".")
		}
		if in.StartIP != nil {
			start = *in.StartIP
		}
		if in.EndIP != nil {
			end = *in.EndIP
		}
	}
	if in.ScanTimeoutMs != nil {
		scanTimeoutMs = *in.ScanTimeoutMs
	}
	if in.HttpTimeoutMs != nil {
		httpTimeoutMs = *in.HttpTimeoutMs
	}

	// Validate: refuse to produce a ScanConfig without a usable IP range
	if subnet == "" || start <= 0 || end <= 0 {
		return scanner.ScanConfig{}, fmt.Errorf(
			"no valid scan range configured: subnet=%q start=%d end=%d — provide a config file or an IPRange filter",
			subnet, start, end,
		)
	}

	// Sanity-checks
	if start < 1 {
		start = 1
	}
	if end > 254 {
		end = 254
	}
	if start > end {
		start, end = end, start
	}

	// Layer 3: hardcoded defaults
	if scanTimeoutMs <= 0 {
		scanTimeoutMs = 300000 // 5 minutes
	}
	if httpTimeoutMs <= 0 {
		httpTimeoutMs = 2000 // 2 seconds
	}

	return scanner.ScanConfig{
		Subnet:      subnet,
		StartIP:     start,
		EndIP:       end,
		ScanTimeout: time.Duration(scanTimeoutMs) * time.Millisecond,
		HTTPTimeout: time.Duration(httpTimeoutMs) * time.Millisecond,
	}, nil
}
