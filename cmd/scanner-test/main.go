package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"shelly-asset-link/scanner"
)

func main() {
	// eigenen Context mit Timeout erzeugen
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg := scanner.ScanConfig{
		Subnet:      "10.0.0",
		StartIP:     160,
		EndIP:       240,
		Timeout:     2 * time.Second,
		MaxParallel: 20,
	}

	// neuer Aufruf mit Context
	results, err := scanner.RunScan(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Scanner-Ergebnisse:")
	fmt.Println("Name\tIP\tMAC\tFirmware")

	for _, d := range results {
		name := safeStr(d.DeviceInfo.Name)
		if name == "" {
			name = "unbenannt"
		}
		fmt.Printf("%s\t%s\t%s\t%s\n",
			name,
			d.IPAddress,
			d.DeviceInfo.Mac,
			d.DeviceInfo.Ver,
		)
	}
}

// Hilfsfunktion für *string
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
