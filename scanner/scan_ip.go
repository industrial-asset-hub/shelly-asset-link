/*
 * SPDX-FileCopyrightText: 2026 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 *
 */

package scanner

import (
    "context"
    "fmt"
    "net"
    "time"
)

// ScanIPResult represents the structured result of scanning a single IP address.
// Fields will be added in later commits when RPC responses are integrated.
type ScanIPResult struct {
    IP string `json:"ip"`
    // Additional fields will follow in Commit 2 and 3
}

// ScanIP performs an RPC-based scan for exactly one IP address.
// This function will replace the old loop-based scanning logic in shellyscanner.go.
func ScanIP(ctx context.Context, ip string, timeout time.Duration) (*ScanIPResult, error) {

    // Debug: start of scan
    fmt.Printf("[ScanIP] Starting scan for %s\n", ip)

    // Validate IP format
    if net.ParseIP(ip) == nil {
        return nil, fmt.Errorf("invalid IP address: %s", ip)
    }

    // TODO: RPC calls will be added in Commit 2
    // TODO: Interface parsing will be added in Commit 3

    // Placeholder result until RPC logic is implemented
    result := &ScanIPResult{
        IP: ip,
    }

    // Debug: end of scan
    fmt.Printf("[ScanIP] Finished: %s\n", ip)

    return result, nil
}
