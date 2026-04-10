/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

// Package mapping provides pure, stateless transformation functions that convert
// Shelly API responses into IAH base schema v0.12.0 device objects.
//
// The single public entry point is Map(). All build* functions are package-private
// and are independently unit-testable without any HTTP or I/O dependencies.
package mapping

import (
	"github.com/industrial-asset-hub/asset-link-sdk/v3/model"

	"shelly-asset-link/shelly"
)

// Map transforms the two Shelly API responses into a fully populated IAH DeviceInfo
// object ready for ConvertToDiscoveredDevice() and subsequent publishing.
//
// Parameters:
//   - info:   validated DeviceInfo from Shelly.GetDeviceInfo (provided by the scanner)
//   - comps:  enriched ComponentsResponse from Shelly.GetComponents (fetched by the handler)
//   - scanIP: the IP address from which GetDeviceInfo responded (used as IP fallback)
//
// The function is pure: no I/O, no side effects, no shared state.
func Map(info shelly.DeviceInfo, comps shelly.ComponentsResponse, scanIP string) *model.DeviceInfo {
	device := model.NewDevice("EthernetDevice", deviceName(info))

	buildIdentifiers(device, info, comps, scanIP)
	buildConnections(device, info, comps, scanIP)
	buildSoftware(device, info)
	buildFunctionalParts(device, comps)
	buildOperations(device, comps)
	buildAnnotations(device, comps)

	return device
}

// deviceName returns the user-defined device name or falls back to "Shelly <model>".
func deviceName(info shelly.DeviceInfo) string {
	if info.Name != nil && *info.Name != "" {
		return *info.Name
	}
	if info.Model != "" {
		return "Shelly " + info.Model
	}
	return "Shelly Device"
}
