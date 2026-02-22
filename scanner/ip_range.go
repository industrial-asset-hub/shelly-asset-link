package scanner

import (
    "fmt"
    "net"
)

func generateIPRange(startIP, endIP string) ([]string, error) {

    start := net.ParseIP(startIP)
    end := net.ParseIP(endIP)

    if start == nil || end == nil {
        return nil, fmt.Errorf("invalid IP")
    }

    var ips []string
    cur := start.To4()

    for ; !cur.Equal(end.To4()); incrementIP(cur) {
        ips = append(ips, cur.String())
    }

    ips = append(ips, endIP)
    return ips, nil
}

func incrementIP(ip net.IP) {
    for i := len(ip) - 1; i >= 0; i-- {
        ip[i]++
        if ip[i] != 0 {
            break
        }
    }
}
