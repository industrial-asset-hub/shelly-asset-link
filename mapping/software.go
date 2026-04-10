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
// The firmware version is sourced from DeviceInfo.fw_id (GetDeviceInfo), which is
// always present and more reliable than sys.config.device.fw_id from GetComponents,
// whose sys block may be absent in some device responses.
func buildSoftware(d *model.DeviceInfo, info shelly.DeviceInfo) {
	if info.FwID == "" {
		return
	}
	d.AddSoftware("firmware", info.FwID, true)
}
