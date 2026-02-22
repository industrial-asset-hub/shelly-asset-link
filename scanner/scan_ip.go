/*
 * SPDX-FileCopyrightText: 2026 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 *
 */

package scanner

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net"
    "net/http"
    "time"
)

// ScanIPResult holds raw RPC responses for a single IP.
// Parsing into structured interfaces will be done in Commit C.
type ScanIPResult struct {
    IP              string          `json:"ip"`
    DeviceInfoRaw   json.RawMessage `json:"deviceInfoRaw,omitempty"`
    WifiConfigRaw   json.RawMessage `json:"wifiConfigRaw,omitempty"`
    WifiStatusRaw   json.RawMessage `json:"wifiStatusRaw,omitempty"`
    EthConfigRaw    json.RawMessage `json:"ethConfigRaw,omitempty"`
    EthStatusRaw    json.RawMessage `json:"ethStatusRaw,omitempty"`
}

// callRPC sends a Shelly RPC request with SPX header and returns raw JSON.
func callRPC(ctx context.Context, client *http.Client, ip string, method string, spxAuth string) (json.RawMessage, error) {

    // Build RPC request body
    body := map[string]any{
        "id":     1,
        "method": method,
    }

    payload, err := json.Marshal(body)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal RPC body: %w", err)
    }

    url := fmt.Sprintf("http://%s/rpc", ip)

    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
    if err != nil {
        return nil, fmt.Errorf("failed to create RPC request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("SPX-Auth", spxAuth)

    // Debug
    fmt.Printf("[ScanIP] RPC → %s: %s\n", ip, method)

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("RPC request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("RPC returned status %d", resp.StatusCode)
    }

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read RPC response: %w", err)
    }

    // Debug
    fmt.Printf("[ScanIP] RPC ← %s: %s OK\n", ip, method)

    return data, nil
}

// ScanIP performs all RPC calls for a single IP and stores raw JSON responses.
func ScanIP(ctx context.Context, ip string, timeout time.Duration, spxAuth string) (*ScanIPResult, error) {

    fmt.Printf("[ScanIP] Starting scan for %s\n", ip)

    if net.ParseIP(ip) == nil {
        return nil, fmt.Errorf("invalid IP address: %s", ip)
    }

    client := &http.Client{
        Timeout: timeout,
    }

    result := &ScanIPResult{
        IP: ip,
    }

    // --- Shelly.GetDeviceInfo ---
    deviceInfo, err := callRPC(ctx, client, ip, "Shelly.GetDeviceInfo", spxAuth)
    if err != nil {
        return nil, fmt.Errorf("device info RPC failed: %w", err)
    }
    result.DeviceInfoRaw = deviceInfo

    // --- Wifi.GetConfig ---
    wifiConfig, err := callRPC(ctx, client, ip, "Wifi.GetConfig", spxAuth)
    if err == nil {
        result.WifiConfigRaw = wifiConfig
    }

    // --- Wifi.GetStatus ---
    wifiStatus, err := callRPC(ctx, client, ip, "Wifi.GetStatus", spxAuth)
    if err == nil {
        result.WifiStatusRaw = wifiStatus
    }

    // --- Eth.GetConfig (optional) ---
    ethConfig, err := callRPC(ctx, client, ip, "Eth.GetConfig", spxAuth)
    if err == nil {
        result.EthConfigRaw = ethConfig
    }

    // --- Eth.GetStatus (optional) ---
    ethStatus, err := callRPC(ctx, client, ip, "Eth.GetStatus", spxAuth)
    if err == nil {
        result.EthStatusRaw = ethStatus
    }

    fmt.Printf("[ScanIP] Finished scan for %s\n", ip)

    return result, nil
}
