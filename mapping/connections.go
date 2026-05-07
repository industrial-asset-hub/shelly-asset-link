/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package mapping

import (
	"github.com/industrial-asset-hub/asset-link-sdk/v3/model"

	"shelly-asset-link/shelly"
)

// buildConnections populates the connection_points list on the IAH device.
//
// It creates:
//   - An EthernetPort for the wired interface (when eth component is present),
//     using eth.config.mac. This also auto-creates the MacIdentifier via AddNic.
//   - An EthernetPort for the WiFi interface (when no eth but wifi is present),
//     using info.Mac from GetDeviceInfo as the MAC source.
//   - An Ipv4Connectivity linked to the EthernetPort, using the primary IP.
func buildConnections(d *model.DeviceInfo, info shelly.DeviceInfo, comps shelly.ComponentsResponse, scanIP string) {
	ip := primaryIP(comps, scanIP)

	if comps.Eth != nil && comps.Eth.Config.Mac != "" {
		// Wired Ethernet: MAC from eth.config — also creates MacIdentifier automatically.
		nicID := d.AddNic("eth", formatMACAddress(comps.Eth.Config.Mac))
		if ip != "" {
			d.AddIPv4(nicID, ip, "255.255.255.0", "")
		}
	} else if info.Mac != "" {
		// WiFi-only or fallback: use the device MAC from GetDeviceInfo.
		nicID := d.AddNic("wifi", formatMACAddress(info.Mac))
		if ip != "" {
			d.AddIPv4(nicID, ip, "255.255.255.0", "")
		}
	}
}
