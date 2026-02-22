package scanner

func mapScanResultToAsset(result *ScanIPResult) any {
    return map[string]any{
        "ip":                result.IP,
        "networkInterfaces": result.NetworkInterfaces,
    }
}

