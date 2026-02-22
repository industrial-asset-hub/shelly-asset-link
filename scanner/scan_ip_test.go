/*
 * SPDX-FileCopyrightText: 2026 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 *
 */

package scanner

import (
    "encoding/json"
    "testing"
)

func TestParseInterfaces(t *testing.T) {

    // --- Fake DeviceInfo ---
    deviceInfo := `{
        "model": "Shelly Plus 1PM",
        "mac": "AA:BB:CC:DD:EE:FF",
        "fw_version": "1.2.3"
    }`

    // --- Fake Wifi.GetConfig ---
    wifiConfig := `{
        "config": {
            "sta":  { "ssid": "OfficeWiFi", "enable": true },
            "sta1": { "ssid": "BackupWiFi", "enable": true }
        }
    }`

    // --- Fake Wifi.GetStatus ---
    wifiStatus := `{
        "status": {
            "sta_ip":  "192.168.10.50",
            "sta1_ip": "192.168.20.50",
            "rssi": -61
        }
    }`

    // --- Fake Eth.GetConfig ---
    ethConfig := `{
        "config": {
            "enable": true,
            "ipv4": {
                "method": "static",
                "ip": "192.168.1.50",
                "netmask": "255.255.255.0",
                "gw": "192.168.1.1"
            }
        }
    }`

    // --- Fake Eth.GetStatus ---
    ethStatus := `{
        "status": {
            "link": true,
            "speed": 100,
            "duplex": "full"
        }
    }`

    // Build ScanIPResult
    result := &ScanIPResult{
        IP:             "192.168.1.50",
        DeviceInfoRaw:  json.RawMessage(deviceInfo),
        WifiConfigRaw:  json.RawMessage(wifiConfig),
        WifiStatusRaw:  json.RawMessage(wifiStatus),
        EthConfigRaw:   json.RawMessage(ethConfig),
        EthStatusRaw:   json.RawMessage(ethStatus),
    }

    // Execute parsing
    if err := parseInterfaces(result); err != nil {
        t.Fatalf("parseInterfaces failed: %v", err)
    }

    if len(result.NetworkInterfaces) != 3 {
        t.Fatalf("expected 3 interfaces, got %d", len(result.NetworkInterfaces))
    }

    // Validate STA
    sta := result.NetworkInterfaces[0]
    if sta.Type != "wifi" || sta.Name != "sta" {
        t.Errorf("STA interface incorrect: %+v", sta)
    }
    if sta.IPv4Address != "192.168.10.50" {
        t.Errorf("STA IP incorrect: %s", sta.IPv4Address)
    }

    // Validate STA1
    sta1 := result.NetworkInterfaces[1]
    if sta1.Type != "wifi" || sta1.Name != "sta1" {
        t.Errorf("STA1 interface incorrect: %+v", sta1)
    }
    if sta1.IPv4Address != "192.168.20.50" {
        t.Errorf("STA1 IP incorrect: %s", sta1.IPv4Address)
    }

    // Validate Ethernet
    eth := result.NetworkInterfaces[2]
    if eth.Type != "ethernet" || eth.Name != "eth0" {
        t.Errorf("Ethernet interface incorrect: %+v", eth)
    }
    if eth.IPv4Address != "192.168.1.50" {
        t.Errorf("Ethernet IP incorrect: %s", eth.IPv4Address)
    }
    if !eth.Link {
        t.Errorf("Ethernet link should be true")
    }
}
