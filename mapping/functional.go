/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package mapping

import (
	"time"

	"github.com/google/uuid"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/model"

	"shelly-asset-link/shelly"
)

// buildFunctionalParts iterates over all functional components (switch:N, cover:N, light:N)
// from GetComponents and creates one Asset per component in the IAH device's
// functional_parts list.
//
// model.Asset is used (instead of model.FunctionalObject) because the IAH schema
// requires a management_state field on every functional_parts entry, and Asset
// is the SDK type that carries this mandatory field.
//
// The name of each entry comes from the user-defined component config.name
// (e.g. "Heckenbewässerung"). If no name is set, the component key (e.g. "switch:0")
// is used as a fallback.
func buildFunctionalParts(d *model.DeviceInfo, comps shelly.ComponentsResponse) {
	if len(comps.Functional) == 0 {
		return
	}

	stateValue := model.ManagementStateValuesUnknown
	foType := model.AssetFunctionalObjectTypeAsset

	for _, fc := range comps.Functional {
		name := fc.Config.Name
		if name == "" {
			name = fc.Key // fallback: "switch:0", "cover:1", etc.
		}

		ts := time.Now().UTC()
		mgmtState := model.ManagementState{
			StateValue:     &stateValue,
			StateTimestamp: &ts,
		}

		n := name // capture loop variable
		fo := model.Asset{
			FunctionalObjectType: &foType,
			Id:                   uuid.New().String(),
			Name:                 &n,
			ManagementState:      mgmtState,
		}
		d.FunctionalParts = append(d.FunctionalParts, fo)
	}
}
