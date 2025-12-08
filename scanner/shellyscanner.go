/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 *
 */

package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// Data node returned by RPC/Shelly.GetDeviceInfo
type ShellyDeviceInfo struct {
	Name       *string `json:"name"`
	ID         string  `json:"id"`
	Mac        string  `json:"mac"`
	Slot       int     `json:"slot"`
	Model      string  `json:"model"`
	Gen        int     `json:"gen"`
	FwID       string  `json:"fw_id"`
	Ver        string  `json:"ver"`
	App        string  `json:"app"`
	AuthEn     bool    `json:"auth_en"`
	AuthDomain *string `json:"auth_domain"`
}

// Aggregated result per device
type ShellyCombined struct {
	DeviceInfo ShellyDeviceInfo       `json:"device_info"`
	Status     map[string]interface{} `json:"status"`
	IPAddress  string                 `json:"ip_address"` // added field
}

// Configuration for the scanner
type ScanConfig struct {
	Subnet      string        // e.g. "10.0.0" without trailing dot
	StartIP     int           // beginning of the IP range
	EndIP       int           // end of the IP range
	Timeout     time.Duration // HTTP timeout
	MaxParallel int           // maximum concurrency
	Logger      *log.Logger   // logger for output
}

// BuildIP constructs an IPv4 address from a subnet string and host number.
// Example: subnet="10.0.0", host=42 → "10.0.0.42"
func BuildIP(subnet string, host int) (string, error) {
	ip := net.ParseIP(fmt.Sprintf("%s.%d", subnet, host))
	if ip == nil {
		return "", fmt.Errorf("invalid IP: %s.%d", subnet, host)
	}
	return ip.String(), nil
}

// RunScan performs device discovery in three steps:
// 1. Call .../rpc/Shelly.GetDeviceInfo; only if successful,
// 2. Call .../rpc/Shelly.GetStatus, then
// 3. Return the combined result.
// The context allows the scan to be cancelled externally.
func RunScan(ctx context.Context, cfg ScanConfig) ([]ShellyCombined, error) {
	// Use fallback logger if none is provided
	logger := cfg.Logger
	if logger == nil {
		logger = log.New(os.Stdout, "scanner: ", log.LstdFlags)
	}

	client := &http.Client{Timeout: cfg.Timeout}

	var results []ShellyCombined
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.MaxParallel)

	// Start stopwatch
	start := time.Now()

	for i := cfg.StartIP; i <= cfg.EndIP; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			ip, err := BuildIP(cfg.Subnet, i)
			if err != nil {
				logger.Printf("Failed to build IP: %v", err)
				return
			}
			logger.Printf("Scanning %s...", ip)

			// Step 1: Request DeviceInfo with context
			urlInfo := fmt.Sprintf("http://%s/rpc/Shelly.GetDeviceInfo", ip)
			reqInfo, _ := http.NewRequestWithContext(ctx, "GET", urlInfo, nil)
			respInfo, err := client.Do(reqInfo)
			if err != nil {
				logger.Printf("No response from %s: %v", ip, err)
				return
			}
			bodyInfo, _ := io.ReadAll(respInfo.Body)
			respInfo.Body.Close()

			var info ShellyDeviceInfo
			if err := json.Unmarshal(bodyInfo, &info); err != nil {
				logger.Printf("Invalid DeviceInfo from %s: %v", ip, err)
				return
			}

			logger.Printf("Found: %s (%s)", safeStr(info.Name), info.Mac)

			// Step 2: Request status only if DeviceInfo succeeded
			status := make(map[string]interface{})
			urlStatus := fmt.Sprintf("http://%s/rpc/Shelly.GetStatus", ip)
			reqStatus, _ := http.NewRequestWithContext(ctx, "GET", urlStatus, nil)
			respStatus, err := client.Do(reqStatus)
			if err == nil {
				bodyStatus, _ := io.ReadAll(respStatus.Body)
				respStatus.Body.Close()

				var raw map[string]json.RawMessage
				if err := json.Unmarshal(bodyStatus, &raw); err == nil {
					for key, val := range raw {
						var node interface{}
						if err := json.Unmarshal(val, &node); err == nil {
							status[key] = node
						}
					}
				}
			} else {
				logger.Printf("Status not available from %s: %v", ip, err)
			}

			// Always record the device, even if status is missing
			mu.Lock()
			results = append(results, ShellyCombined{
				DeviceInfo: info,
				Status:     status,
				IPAddress:  ip,
			})
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Compute statistics
	elapsed := time.Since(start)
	totalIPs := cfg.EndIP - cfg.StartIP + 1
	found := len(results)

	// Count Shelly devices (heuristic: App or Model is set)
	shellyCount := 0
	for _, r := range results {
		if r.DeviceInfo.App != "" || r.DeviceInfo.Model != "" {
			shellyCount++
		}
	}

	// Summary output
	logger.Printf("%d IP addresses scanned in %02d:%02d, %d devices detected, %d identified as Shelly",
		totalIPs,
		int(elapsed.Minutes()),
		int(elapsed.Seconds())%60,
		found,
		shellyCount,
	)

	return results, nil
}

// Helper: safely dereference *string
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
