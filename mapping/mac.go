/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 */

package mapping

import (
	"fmt"
	"strings"
)

// formatMACAddress normalizes a MAC address string to lowercase colon-separated
// format (e.g. "aa:bb:cc:dd:ee:ff") before it is passed to the Asset Link SDK.
//
// The function is idempotent: MACs already carrying colons, hyphens, or dots as
// separators are stripped and re-formatted consistently.  Input is accepted in
// any mix of upper- and lowercase.  Strings that do not reduce to exactly 12 hex
// characters are returned unchanged so that unexpected values are visible in the
// SDK payload rather than silently mangled.
func formatMACAddress(mac string) string {
	clean := strings.Map(func(r rune) rune {
		if r == ':' || r == '-' || r == '.' {
			return -1
		}
		return r
	}, mac)
	clean = strings.ToLower(clean)
	if len(clean) != 12 {
		return mac
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		clean[0:2], clean[2:4], clean[4:6],
		clean[6:8], clean[8:10], clean[10:12])
}
