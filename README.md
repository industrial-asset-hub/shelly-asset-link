# Shelly Asset Link

**Shelly Asset Link** is a Go service that discovers [Shelly](https://www.shelly.com/collections/all-products) Gen2 and later IoT devices on a local network and registers them as digital assets in the [Siemens Industrial Asset Hub](https://industrial-assets.io). It implements the Discovery interface of the [Asset Link SDK v3](https://github.com/industrial-asset-hub/asset-link-SDK) and communicates with the IAH via gRPC.

> **Disclaimer:** This is a private project and not an official product of Siemens AG or Shelly SE.

---

## Features

- IP range scan using Shelly's HTTP RPC API (`Shelly.GetDeviceInfo`, `Shelly.GetComponents`)
- Automatic device model detection (Gen2 Plus/Pro, Gen3)
- Populates IAH nameplate, MAC/IP identifiers, software artifacts, functional parts (switch, cover, light channels), and operational metadata
- Configurable scan range, scan timeout, and per-request HTTP timeout
- Two-level configuration priority: IAH discovery request > config file 
- Runs as a minimal Docker container (scratch-based, multi-arch: `linux/amd64`, `linux/arm64`)

## Supported Devices

Shelly Asset Link supports all devices running the Gen2 and later RPC firmware (ESP32-based). These devices respond to the `Shelly.GetDeviceInfo` and `Shelly.GetComponents` RPC endpoints over HTTP.

**Supported:**

- Shelly Plus series (Gen2): Plus 1, Plus 1PM, Plus 2PM, Plus i4, Plus Plug S, Plus HT, Plus Smoke
- Shelly Pro series (Gen2, DIN-rail): Pro 1, Pro 1PM, Pro 2, Pro 2PM, Pro 3, Pro 3EM, Pro 4PM
- Shelly Gen3 series: 1 Gen3, 1PM Gen3, 2PM Gen3, Plug Gen3, Dimmer Gen3
- Shelly Gen4 series:  1 etc.

**Not supported:**

- Gen1 devices (Shelly 1, 2.5, Dimmer 2, RGBW2, older Plug S) — these use a legacy REST API and do not expose RPC endpoints.

## Prerequisites

- Docker and Docker Compose
- A running [Asset-Gateway](https://github.com/industrial-asset-hub/asset-gateway) container on the same Docker network
- Shelly Asset Link must be registered with the Asset-Gateway (use [al-ctl](https://github.com/industrial-asset-hub/asset-link-sdk) for registration and status checks)
- For production use with the Industrial Asset Hub: an IAH subscription and a server user with API credentials

## Installation

```bash
# Create the config directory
mkdir -p config

# Place your configuration file (see Configuration section)
cp shelly_al_config.json config/

# Start the service
docker compose up --force-recreate
```

The `docker-compose.yml` builds the image locally and connects to the external Docker network `cdm` (alias `01-cdm`), which is shared with the Asset-Gateway. The gRPC server listens on port `8081` inside the container.

To expose the gRPC port on the host for local testing, uncomment the `ports` section in `docker-compose.yml`.

## Configuration

### Config file

The service reads scan parameters from a JSON file mounted into the container at the path specified by the `AL_CONFIG_PATH` environment variable (default: `/config/shelly_al_config.json`).

```json
{
  "subnet": "10.0.0",
  "startIP": 1,
  "endIP": 254,
  "scanTimeoutMs": 300000,
  "httpTimeoutMs": 2000
}
```

| Field           | Description                                                         | Default  |
|-----------------|---------------------------------------------------------------------|----------|
| `subnet`        | Base network segment to scan, without trailing dot (e.g. `10.0.20`) | `10.0.0` |
| `startIP`       | First host octet to scan                                            | `1`      |
| `endIP`         | Last host octet to scan                                             | `254`    |
| `scanTimeoutMs` | Maximum total duration for the entire scan in milliseconds          | `300000` |
| `httpTimeoutMs` | Per-request HTTP timeout for individual Shelly API calls in milliseconds | `2000` |

### Precedence

1. Discovery request parameters passed by the IAH (filters and options) — highest priority
2. Values in `shelly_al_config.json`

### Supported discovery filters

| Key       | Type   | Description                                           |
|-----------|--------|-------------------------------------------------------|
| `IPRange` | string | IP range to scan, e.g. `10.0.0.1-254`              |

### Supported discovery options

| Key             | Type   | Description                                                |
|-----------------|--------|------------------------------------------------------------|
| `scanTimeoutMs` | string | Maximum total scan duration in milliseconds                |
| `httpTimeoutMs` | string | Per-request HTTP timeout for Shelly API calls in milliseconds |

## Usage

### Industrial Asset Hub UI

Open the IAH Inbox, start a scan, select the Asset-Gateway, choose Shelly Asset Link as the scanner, and fill in the IP range to scan. Discovered devices appear in the Inbox while discovery job is running.

### API

To start a discovery job via the IAH REST API:

1. Set up a server user in [Siemens Xcelerator Admin Console](https://cloud.sw.siemens.com/admin).

2. Obtain an access token:

   ```bash
   export TOKEN="$(
     curl -s -X POST "https://iahprod.eu1.sws.siemens.com/oauth/token" \
       -H "Content-Type: application/x-www-form-urlencoded" \
       -d "grant_type=client_credentials&client_id=<client_id>&client_secret=<client_secret>" \
     | jq -r '.access_token')"
   ```

3. Start a discovery job:

   ```bash
   curl -sS -X POST \
     "https://cloud.eu1.sws.siemens.com/api/assethubapi/v1/discovery/v1-earlyaccess/gateways/<gateway-id>/discoveries" \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "assetLinkId": "cdm-device-class-driver-mlp0911.cdm.al.shelly",
       "deviceClass": "mlp0911.cdm.al.shelly",
       "scanMode": "FULL",
       "jobType": "DISCOVERY",
       "discoveryFilters": [
         {"key": "IPRange", "comparisonOperator": "EQUAL", "value": "10.0.0.1-10.0.0.254"}
       ],
       "discoveryOptions": [
         {"key": "scanTimeoutMs", "value": "300000"},
         {"key": "httpTimeoutMs", "value": "2000"}
       ]
     }'
   ```

For further information on managing the Inbox, onboarding devices, and using IAH services, refer to the [Industrial Asset Hub Documentation](https://industrial-assets.io).

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## Resources

- [Asset Link SDK](https://github.com/industrial-asset-hub/asset-link-SDK)
- [Industrial Asset Hub](https://industrial-assets.io)
- [Asset-Gateway](https://github.com/industrial-asset-hub/asset-gateway)
- [al-ctl / Asset Link SDK CLI](https://github.com/industrial-asset-hub/asset-link-sdk)
- [Siemens Xcelerator Admin Console](https://cloud.sw.siemens.com/admin)
- [Shelly Group SE](https://www.shelly.com/pages/location-selector)
- [Shelly Portfolio](https://www.shelly.com/collections/all-products)
- [Shelly Developer API Documentation](https://shelly-api-docs.shelly.cloud/)
