/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 *
 */

package shellycfg

import (
	"net"
	"strconv"
	"strings"
)

func ParseCIDR(cidr string) (subnet string, start, end int, ok bool) {
	ip, ipnet, err := net.ParseCIDR(strings.TrimSpace(cidr))
	if err != nil || ip == nil || ipnet == nil {
		return "", 0, 0, false
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return "", 0, 0, false
	}
	ones, bits := ipnet.Mask.Size()
	if bits != 32 {
		return "", 0, 0, false
	}
	network := ip4.Mask(ipnet.Mask)
	start, end = 1, 254
	if ones > 24 {
		hostBits := 32 - ones
		hostCount := (1 << hostBits) - 2
		if hostCount < 2 {
			hostCount = 2
		}
		start, end = 1, hostCount
	}
	subnet = strings.Join(strings.Split(network.String(), ".")[:3], ".")
	return subnet, start, end, true
}

func ParseIPRange(r string) (subnet string, start, end int, ok bool) {
	s := strings.TrimSpace(r)
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return "", 0, 0, false
	}
	ipA := net.ParseIP(strings.TrimSpace(parts[0]))
	ipB := net.ParseIP(strings.TrimSpace(parts[1]))
	if ipA == nil || ipB == nil {
		return "", 0, 0, false
	}
	a4, b4 := ipA.To4(), ipB.To4()
	if a4 == nil || b4 == nil {
		return "", 0, 0, false
	}
	if a4[0] != b4[0] || a4[1] != b4[1] || a4[2] != b4[2] {
		return "", 0, 0, false
	}
	subnet = strings.Join([]string{strconv.Itoa(int(a4[0])), strconv.Itoa(int(a4[1])), strconv.Itoa(int(a4[2]))}, ".")
	start, end = int(a4[3]), int(b4[3])
	if start > end {
		start, end = end, start
	}
	if start < 1 {
		start = 1
	}
	if end > 254 {
		end = 254
	}
	return subnet, start, end, true
}
