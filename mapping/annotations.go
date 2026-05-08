/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package mapping

import (
	"strconv"

	"github.com/industrial-asset-hub/asset-link-sdk/v3/model"

	"shelly-asset-link/shelly"
)

// buildAnnotations populates the instance_annotations list on the IAH device with
// low-volatility, contextually important key-value pairs.
//
// Intentionally excluded: uptime, counters, real-time power readings, and other
// high-frequency time-series data that should be observed via monitoring systems
// rather than stored as asset metadata.
//
// Included:
//   - device.name              — user-defined device name
//   - device.profile           — active device profile (when set)
//   - connectivity.bluetooth.enabled
//   - connectivity.mqtt.enabled
//   - diagnostics.restartRequired
//   - firmware.availableVersion.<channel>  — e.g. "stable", "beta" (when updates available)
func buildAnnotations(d *model.DeviceInfo, comps shelly.ComponentsResponse) {
	add := func(key, value string) {
		if value == "" {
			return
		}
		k, v := key, value
		d.InstanceAnnotations = append(d.InstanceAnnotations,
			model.InstanceAnnotation{Key: &k, Value: &v})
	}

	if comps.Sys != nil {
		add("device.name", comps.Sys.Config.Device.Name)
		add("device.profile", comps.Sys.Config.Device.Profile)
		add("diagnostics.restartRequired",
			strconv.FormatBool(comps.Sys.Status.RestartRequired))

		for channel, upd := range comps.Sys.Status.AvailableUpdates {
			add("firmware.availableVersion."+channel, upd.Version)
		}
	}

	if comps.Bt != nil {
		add("connectivity.bluetooth.enabled",
			strconv.FormatBool(comps.Bt.Config.Enable))
	}

	if comps.Mqtt != nil {
		add("connectivity.mqtt.enabled",
			strconv.FormatBool(comps.Mqtt.Config.Enable))
	}
}
