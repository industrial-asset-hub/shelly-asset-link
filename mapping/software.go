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

// buildSoftware adds exactly one SoftwareArtifact to the IAH device representing
// the currently installed firmware.
//
// The firmware version is sourced from sys.config.device.fw_id (GetComponents),
// which is the authoritative runtime source — not DeviceInfo.fw_id.
func buildSoftware(d *model.DeviceInfo, comps shelly.ComponentsResponse) {
	if comps.Sys == nil {
		return
	}
	fwID := comps.Sys.Config.Device.FwID
	if fwID == "" {
		return
	}
	d.AddSoftware("firmware", fwID, true)
}
