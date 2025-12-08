/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package shellycfg

import (
	"strings"
	"time"

	"shelly-asset-link/scanner"
)

// BuildFromInputs merges inputs delivered at runtime (Discovery-filters/options)
// w/ the static file config (fileCfg) and delivers effective ScanConfig
// for shellyscanner

// new BuildFromInputs w/ new order and priority

func BuildFromInputs(in Inputs, fileCfg *FileScanConfig) scanner.ScanConfig {
	// Defaults (Fallback)
	subnet, start, end := "10.0.0", 1, 254
	timeoutMs, maxPar := 300000, 10

	// use file cfg if existing
	if fileCfg != nil {
		subnet, start, end = fileCfg.Subnet, fileCfg.StartIP, fileCfg.EndIP
		timeoutMs, maxPar = fileCfg.TimeoutMs, fileCfg.MaxParallel
	}

	// --- Filter-logic only ipRange for a start ---
	s := strings.TrimSpace(in.IpRange)
	if s != "" {
		if sn, st, en, ok := ParseIPRange(s); ok {
			subnet, start, end = sn, st, en
		}
		// Parse-error → fileCfg/Defaults stays active
	} else {
		// explicit Subnet/Start/End from Inputs
		if strings.TrimSpace(in.Subnet) != "" {
			subnet = strings.TrimSuffix(strings.TrimSpace(in.Subnet), ".")
		}
		if in.StartIP != nil {
			start = *in.StartIP
		}
		if in.EndIP != nil {
			end = *in.EndIP
		}
		// CIDR & Options ignored for now
	}

	// --- Sanity-Checks ---
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
	if maxPar <= 0 {
		maxPar = 10
	}

	return scanner.ScanConfig{
		Subnet:      subnet,
		StartIP:     start,
		EndIP:       end,
		Timeout:     time.Duration(timeoutMs) * time.Millisecond,
		MaxParallel: maxPar,
	}
}
