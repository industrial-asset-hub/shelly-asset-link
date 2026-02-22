//go:build manual
// +build manual

package scanner

import (
    "context"
    "testing"
    "time"
)

type TestPublisher struct {
    t *testing.T
}

func (p *TestPublisher) PublishDiscovery(asset any) error {
    p.t.Logf("DISCOVERY: %+v\n", asset)
    return nil
}

func TestManualScan(t *testing.T) {

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    startIP := "10.0.20.20"
    endIP   := "10.0.20.130"
    timeout := 2 * time.Second

    publisher := &TestPublisher{t: t}

    if err := runStreamingScan(ctx, startIP, endIP, timeout, publisher); err != nil {
        t.Fatalf("Scan failed: %v", err)
    }
}
