/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package mapping

import (
	"fmt"

	"github.com/industrial-asset-hub/asset-link-sdk/v3/model"

	"shelly-asset-link/shelly"
)

// buildIdentifiers populates the asset_identifiers list on the IAH device:
//   - ProductSerialIdentifier  (via AddNameplate — manufacturer "Shelly", serial = DeviceInfo.id)
//   - Ipv4Identifier           (primary IP resolved from ComponentsResponse or scanIP fallback)
//
// The MacIdentifier is added automatically by buildConnections via model.AddNic().
func buildIdentifiers(d *model.DeviceInfo, info shelly.DeviceInfo, comps shelly.ComponentsResponse, scanIP string) {
	// 1. ProductSerialIdentifier
	d.AddNameplate(
		"Shelly",                    // manufacturer.name
		"",                          // uriOfTheProduct — not available
		info.Model,                  // manufacturer_product.product_id  e.g. "SNSW-003X16EU"
		info.App,                    // manufacturerProductDesignation    e.g. "Pro3"
		fmt.Sprintf("%d", info.Gen), // hardwareVersion = generation      e.g. "3"
		info.ID,                     // serial_number  e.g. "shellyPro3-a8032ab4c9ec"
	)

	// 2. Ipv4Identifier
	ip := primaryIP(comps, scanIP)
	if ip != "" {
		idType := model.Ipv4IdentifierAssetIdentifierTypeIpv4Identifier
		uncertainty := 2 // IP is more volatile than MAC or serial number
		d.AssetIdentifiers = append(d.AssetIdentifiers, model.Ipv4Identifier{
			AssetIdentifierType:   &idType,
			Ipv4Address:           &ip,
			IdentifierUncertainty: &uncertainty,
		})
	}
}

// primaryIP resolves the primary IP address with the following priority:
//  1. eth.status.ip   — wired Ethernet (most stable)
//  2. wifi.status.sta_ip — WiFi station IP
//  3. scanIP          — fallback: the IP from which GetDeviceInfo responded
func primaryIP(comps shelly.ComponentsResponse, scanIP string) string {
	if comps.Eth != nil && comps.Eth.Status.IP != "" {
		return comps.Eth.Status.IP
	}
	if comps.Wifi != nil && comps.Wifi.Status.StaIP != "" {
		return comps.Wifi.Status.StaIP
	}
	return scanIP
}
