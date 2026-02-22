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

type NetworkInterface struct {
    Type          string `json:"type"`          // wifi or ethernet
    Name          string `json:"name"`          // sta, sta1, eth0
    MAC           string `json:"mac,omitempty"`
    SSID          string `json:"ssid,omitempty"`
    IPv4Address   string `json:"ipv4Address,omitempty"`
    IPv4Netmask   string `json:"ipv4Netmask,omitempty"`
    IPv4Gateway   string `json:"ipv4Gateway,omitempty"`
    RSSI          int    `json:"rssi,omitempty"`
    Link          bool   `json:"link,omitempty"`
    Speed         int    `json:"speed,omitempty"`
    Duplex        string `json:"duplex,omitempty"`
    Connected     bool   `json:"connected"`
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

// Parsing of raw RPC responses into structured NetworkInterface list
func parseInterfaces(result *ScanIPResult) error {

    // --- Parse DeviceInfo for MAC ---
    var devInfo struct {
        MAC string `json:"mac"`
    }
    if result.DeviceInfoRaw != nil {
        _ = json.Unmarshal(result.DeviceInfoRaw, &devInfo)
    }

    // --- Parse Wifi.GetConfig ---
    var wifiConfig struct {
        Config struct {
            STA struct {
                SSID   string `json:"ssid"`
                Enable bool   `json:"enable"`
            } `json:"sta"`
            STA1 struct {
                SSID   string `json:"ssid"`
                Enable bool   `json:"enable"`
            } `json:"sta1"`
        } `json:"config"`
    }
    if result.WifiConfigRaw != nil {
        _ = json.Unmarshal(result.WifiConfigRaw, &wifiConfig)
    }

    // --- Parse Wifi.GetStatus ---
    var wifiStatus struct {
        Status struct {
            STAIP  string `json:"sta_ip"`
            STA1IP string `json:"sta1_ip"`
            RSSI   int    `json:"rssi"`
        } `json:"status"`
    }
    if result.WifiStatusRaw != nil {
        _ = json.Unmarshal(result.WifiStatusRaw, &wifiStatus)
    }

    // --- Parse Eth.GetConfig ---
    var ethConfig struct {
        Config struct {
            Enable bool `json:"enable"`
            IPv4   struct {
                Method  string `json:"method"`
                IP      string `json:"ip"`
                Netmask string `json:"netmask"`
                Gateway string `json:"gw"`
            } `json:"ipv4"`
        } `json:"config"`
    }
    if result.EthConfigRaw != nil {
        _ = json.Unmarshal(result.EthConfigRaw, &ethConfig)
    }

    // --- Parse Eth.GetStatus ---
    var ethStatus struct {
        Status struct {
            Link   bool   `json:"link"`
            Speed  int    `json:"speed"`
            Duplex string `json:"duplex"`
        } `json:"status"`
    }
    if result.EthStatusRaw != nil {
        _ = json.Unmarshal(result.EthStatusRaw, &ethStatus)
    }

    // --- Build interface list ---
    var interfaces []NetworkInterface

    // WiFi STA
    if wifiConfig.Config.STA.Enable {
        interfaces = append(interfaces, NetworkInterface{
            Type:        "wifi",
            Name:        "sta",
            MAC:         devInfo.MAC,
            SSID:        wifiConfig.Config.STA.SSID,
            IPv4Address: wifiStatus.Status.STAIP,
            RSSI:        wifiStatus.Status.RSSI,
            Connected:   wifiStatus.Status.STAIP != "",
        })
    }

    // WiFi STA1
    if wifiConfig.Config.STA1.Enable {
        interfaces = append(interfaces, NetworkInterface{
            Type:        "wifi",
            Name:        "sta1",
            MAC:         devInfo.MAC,
            SSID:        wifiConfig.Config.STA1.SSID,
            IPv4Address: wifiStatus.Status.STA1IP,
            RSSI:        wifiStatus.Status.RSSI,
            Connected:   wifiStatus.Status.STA1IP != "",
        })
    }

    // Ethernet
    if ethConfig.Config.Enable {
        interfaces = append(interfaces, NetworkInterface{
            Type:        "ethernet",
            Name:        "eth0",
            MAC:         devInfo.MAC,
            IPv4Address: ethConfig.Config.IPv4.IP,
            IPv4Netmask: ethConfig.Config.IPv4.Netmask,
            IPv4Gateway: ethConfig.Config.IPv4.Gateway,
            Link:        ethStatus.Status.Link,
            Speed:       ethStatus.Status.Speed,
            Duplex:      ethStatus.Status.Duplex,
            Connected:   ethStatus.Status.Link,
        })
    }

    result.Interfaces = interfaces
    return nil
}

// Parse interfaces from raw RPC data
if err := parseInterfaces(result); err != nil {
    fmt.Printf("[ScanIP] Interface parsing failed for %s: %v\n", ip, err)
}