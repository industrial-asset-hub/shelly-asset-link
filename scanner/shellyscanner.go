/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
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
	"time"

	"shelly-asset-link/shelly"
)

// DiscoveredDevice is emitted by RunScan for every IP that responds successfully
// to GET /rpc/Shelly.GetDeviceInfo. It bundles the confirmed IP address with the
// validated device information — no status, no assumptions about connectivity beyond
// the single successful HTTP call.
type DiscoveredDevice struct {
	IPAddress  string            // IP address at which the device was reached
	DeviceInfo shelly.DeviceInfo // full response from Shelly.GetDeviceInfo
}

// ScanConfig holds the parameters for a single scan run.
type ScanConfig struct {
	Subnet      string        // e.g. "10.0.0" without trailing dot
	StartIP     int           // first host number of the IP range
	EndIP       int           // last host number of the IP range
	ScanTimeout time.Duration // maximum total duration for the entire scan
	HTTPTimeout time.Duration // per-request HTTP timeout for individual Shelly API calls
	Logger      *log.Logger   // optional logger; falls back to os.Stdout
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

// RunScan starts an asynchronous sequential scan over the configured IP range.
// It performs GET /rpc/Shelly.GetDeviceInfo for each IP as a first-stage filter:
// only IPs that return a valid Shelly response are emitted on deviceCh.
//
// Returns two read-only channels:
//   - deviceCh: emits one DiscoveredDevice per confirmed Shelly device, immediately upon detection.
//   - errCh: carries at most one fatal error; closed when the goroutine exits.
//
// The caller must fully drain deviceCh (e.g. via range) to prevent the goroutine from blocking.
// The context can be used to cancel or time out the scan externally.
func RunScan(ctx context.Context, cfg ScanConfig) (<-chan DiscoveredDevice, <-chan error) {
	deviceCh := make(chan DiscoveredDevice, 1)
	errCh := make(chan error, 1)

	go func() {
		defer close(deviceCh)
		defer close(errCh)
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("scanner panic: %v", r)
			}
		}()

		logger := cfg.Logger
		if logger == nil {
			logger = log.New(os.Stdout, "scanner: ", log.LstdFlags)
		}

		client := &http.Client{Timeout: cfg.HTTPTimeout}
		start := time.Now()
		found := 0

		for i := cfg.StartIP; i <= cfg.EndIP; i++ {
			select {
			case <-ctx.Done():
				logger.Printf("Scan aborted after %d IPs: %v", i-cfg.StartIP, ctx.Err())
				return
			default:
			}

			ip, err := BuildIP(cfg.Subnet, i)
			if err != nil {
				logger.Printf("Failed to build IP for host %d: %v", i, err)
				continue
			}

			// First-stage filter: only IPs with a valid GetDeviceInfo response proceed.
			urlInfo := fmt.Sprintf("http://%s/rpc/Shelly.GetDeviceInfo", ip)
			reqInfo, _ := http.NewRequestWithContext(ctx, "GET", urlInfo, nil)
			respInfo, err := client.Do(reqInfo)
			if err != nil {
				continue // no Shelly at this IP — silent skip
			}
			bodyInfo, _ := io.ReadAll(respInfo.Body)
			respInfo.Body.Close()

			var info shelly.DeviceInfo
			if err := json.Unmarshal(bodyInfo, &info); err != nil {
				logger.Printf("Invalid DeviceInfo JSON from %s: %v", ip, err)
				continue
			}

			logger.Printf("Found Shelly device: %s (MAC: %s) at %s", safeStr(info.Name), info.Mac, ip)

			found++
			deviceCh <- DiscoveredDevice{
				IPAddress:  ip,
				DeviceInfo: info,
			}
		}

		elapsed := time.Since(start)
		totalIPs := cfg.EndIP - cfg.StartIP + 1
		logger.Printf(
			"%d IP addresses scanned in %02d:%02d, %d Shelly devices found",
			totalIPs,
			int(elapsed.Minutes()),
			int(elapsed.Seconds())%60,
			found,
		)
	}()

	return deviceCh, errCh
}

// safeStr safely dereferences a *string pointer.
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
