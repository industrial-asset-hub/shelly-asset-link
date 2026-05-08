/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 *
 */

package shellycfg

// Inputs are normalized values extracted in handler.go via SDK getters.
type Inputs struct {
	Cidr          string
	IpRange       string
	Subnet        string
	StartIP       *int
	EndIP         *int
	ScanTimeoutMs *int
	HttpTimeoutMs *int
}
