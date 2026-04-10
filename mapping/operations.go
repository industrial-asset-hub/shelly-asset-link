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

// buildOperations populates the asset_operations list on the IAH device.
//
// Static operations (always present):
//   - reboot         — every Shelly device supports a remote reboot
//   - factory_reset  — every Shelly device supports a factory reset
//
// Dynamic operations (conditional):
//   - firmware_update — activation_flag=true when sys.status.available_updates is non-empty
//   - calibrate       — added only when at least one cover:N component is present
func buildOperations(d *model.DeviceInfo, comps shelly.ComponentsResponse) {
	// Static: always available
	d.AddCapabilities("reboot", true)
	d.AddCapabilities("reset_to_factory", true)

	// Dynamic: firmware update availability
	hasFwUpdate := comps.Sys != nil && len(comps.Sys.Status.AvailableUpdates) > 0
	d.AddCapabilities("firmware_update", hasFwUpdate)

	// Dynamic: calibrate only for devices with a cover (roller/blind) component
	for _, fc := range comps.Functional {
		if fc.Kind == "cover" {
			d.AddCapabilities("calibrate", true)
			break
		}
	}
}
