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
//  1. IPRange filter from the Discovery request
//  2. File config (shelly_al_config.json)
//
// Returns an error when no valid IP range can be determined from the available sources.
func BuildFromInputs(in Inputs, fileCfg *FileScanConfig) (scanner.ScanConfig, error) {
	var subnet string
	var start, end, timeoutMs int

	// Layer 1: seed from file config when available
	if fileCfg != nil {
		subnet = fileCfg.Subnet
		start = fileCfg.StartIP
		end = fileCfg.EndIP
		timeoutMs = fileCfg.TimeoutMs
	}

	// Layer 2: IPRange filter from the Discovery request takes precedence
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
	if timeoutMs <= 0 {
		timeoutMs = 300000
	}

	return scanner.ScanConfig{
		Subnet:  subnet,
		StartIP: start,
		EndIP:   end,
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}, nil
}
