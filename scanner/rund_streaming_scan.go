package scanner

import (
    "context"
    "fmt"
    "time"
)

func runStreamingScan(
    ctx context.Context,
    startIP, endIP string,
    timeout time.Duration,
    publisher Publisher,
) error {

    ips, err := generateIPRange(startIP, endIP)
    if err != nil {
        return fmt.Errorf("invalid IP range: %w", err)
    }

    for _, ip := range ips {

        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        result, err := ScanIP(ctx, ip, timeout)
        if err != nil {
            fmt.Printf("[Scan] %s unreachable: %v\n", ip, err)
            continue
        }

        asset := mapScanResultToAsset(result)

        if err := publisher.PublishDiscovery(asset); err != nil {
            fmt.Printf("[Scan] publish failed: %v\n", err)
        }
    }

    return nil
}
