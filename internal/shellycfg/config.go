/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 *
 */

package shellycfg

import (
	"encoding/json"
	"os"
	"strings"
)

type FileScanConfig struct {
	Subnet    string `json:"subnet"`
	StartIP   int    `json:"startIP"`
	EndIP     int    `json:"endIP"`
	TimeoutMs int    `json:"timeout"`
}

func LoadFileConfig(path string) *FileScanConfig {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var c FileScanConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil
	}
	c.Subnet = strings.TrimSuffix(strings.TrimSpace(c.Subnet), ".")
	if c.TimeoutMs <= 0 {
		c.TimeoutMs = 500
	}
	if c.StartIP <= 0 {
		c.StartIP = 1
	}
	if c.EndIP <= 0 {
		c.EndIP = 254
	}
	return &c
}
