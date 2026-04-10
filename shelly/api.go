/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package shelly

// DeviceInfo represents the response from GET /rpc/Shelly.GetDeviceInfo.
// This call is the first-stage validation performed by the scanner to confirm
// that an IP address belongs to a Shelly device.
type DeviceInfo struct {
	Name       *string `json:"name"`
	ID         string  `json:"id"`
	Mac        string  `json:"mac"`
	Slot       int     `json:"slot"`
	Model      string  `json:"model"`
	Gen        int     `json:"gen"`
	FwID       string  `json:"fw_id"`
	Ver        string  `json:"ver"`
	App        string  `json:"app"`
	AuthEn     bool    `json:"auth_en"`
	AuthDomain *string `json:"auth_domain"`
}

// ComponentsResponse is the fully parsed result of GET /rpc/Shelly.GetComponents.
// All fields are optional — not every device model has every component.
type ComponentsResponse struct {
	Sys        *SysComponent
	Eth        *EthComponent
	Wifi       *WifiComponent
	Bt         *BtComponent
	Mqtt       *MqttComponent
	Functional []FunctionalComponent
}

// SysComponent holds the system configuration and status.
type SysComponent struct {
	Config SysConfig
	Status SysStatus
}

// SysConfig is the configuration block of the sys component.
type SysConfig struct {
	Device SysDeviceConfig `json:"device"`
}

// SysDeviceConfig contains device-level settings persisted in sys.config.device.
type SysDeviceConfig struct {
	Name    string `json:"name"`
	Profile string `json:"profile,omitempty"`
	FwID    string `json:"fw_id"`
}

// SysStatus holds runtime status fields for the sys component.
type SysStatus struct {
	AvailableUpdates map[string]AvailableUpdate `json:"available_updates,omitempty"`
	RestartRequired  bool                       `json:"restart_required"`
}

// AvailableUpdate describes a single available firmware update channel (e.g. "stable", "beta").
type AvailableUpdate struct {
	Version string `json:"version"`
}

// EthComponent holds Ethernet interface configuration and status.
type EthComponent struct {
	Config EthConfig
	Status EthStatus
}

// EthConfig contains the static Ethernet configuration (MAC address).
type EthConfig struct {
	Mac string `json:"mac"`
}

// EthStatus contains the current Ethernet status (assigned IP address).
type EthStatus struct {
	IP string `json:"ip"`
}

// WifiComponent holds WiFi status information.
type WifiComponent struct {
	Status WifiStatus
}

// WifiStatus contains the current WiFi connection status.
type WifiStatus struct {
	StaIP string `json:"sta_ip"`
}

// BtComponent holds Bluetooth configuration.
type BtComponent struct {
	Config BtConfig
}

// BtConfig contains the Bluetooth enable flag.
type BtConfig struct {
	Enable bool `json:"enable"`
}

// MqttComponent holds MQTT configuration.
type MqttComponent struct {
	Config MqttConfig
}

// MqttConfig contains the MQTT enable flag.
type MqttConfig struct {
	Enable bool `json:"enable"`
}

// FunctionalComponent represents a user-visible channel component
// (switch:N, cover:N, light:N). Each such component maps to one
// functional_part in the IAH device model.
type FunctionalComponent struct {
	Key    string          // e.g. "switch:0", "cover:1", "light:2"
	Kind   string          // "switch", "cover", or "light"
	Index  int             // channel index: 0, 1, 2, ...
	Config ComponentConfig // user-facing configuration
}

// ComponentConfig holds the user-defined configuration for a functional component.
type ComponentConfig struct {
	Name string `json:"name"` // user-defined name, e.g. "Heckenbewässerung"
}
