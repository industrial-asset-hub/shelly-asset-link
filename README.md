# Shelly Asset Link


> **⚠️ Disclaimer:**  
> This repository contains a personal project developed using publicly available SDK functions and APIs. 
> It is **not** an official Siemens or Shelly product and is provided under the MIT License.


## Motivation
**shelly-asset-link** is an integration module based on the [Asset Link SDK](https://github.com/industrial-asset-hub/asset-link-SDK) of the [Siemens Industrial Asset Hub](https://industrial-assets.io). It enables the integration of IoT automation components from Shelly Group SE into the Industrial Asset Hub. This allows [Shelly devices](https://www.shelly.com/collections/all-products) to be represented as digital assets within industrial applications and ecosystems, providing following advantages:

- **Interoperability:** The asset link seamlessly connects Shelly IoT automation components with the Siemens Industrial Asset Hub. The integration focuses on instantiated device data, providing a digital nameplate and all relevant descriptive metadata for each Shelly device. This ensures that devices are represented consistently as standardized digital assets within the Industrial Asset Hub ecosystem, enabling the use of Shelly devices in digital twins, asset management, and automation processes.

- **Standardized Interfaces:** The asset link provides plug-and-play integration into the industrial-grade inventory service, Industrial Asset Hub, without the need for custom adapters or proprietary protocols. Built on a future-proof architecture, it enables seamless extension to additional management systems and services, ensuring scalability and long-term interoperability.

## Prerequisites
To run Shelly Asset Link beyond developer testing and use it productively:

- [Asset-Gateway](https://github.com/industrial-asset-hub/asset-gateway) must be running as a container (central registration and communication service).
- Shelly-Asset-Link container must be started (e.g., via docker-compose up).
- Shelly-Asset-Link must be registered with the Asset-Gateway.
- Optional but recommended: [al-ctl](https://github.com/industrial-asset-hub/asset-link-sdk/blob/main/cmd/al-ctl/al-ctl.go) CLI tool for registration and status checks (see [Asset-Link-SDK](https://github.com/industrial-asset-hub/asset-link-sdk)).
- For production use, an Industrial Asset Hub subscription is required to access IAH services and features.

Start the Asset Gateway first and continue with the Shelly-Asset-Link startup by executing the following commands:


```bash
# Create a config folder
mkdir -p config 
# Get compose file for Shelly-Asset-Link
wget https://raw.githubusercontent.com/industrial-asset-hub/shelly-asset-link/refs/heads/main/docker-compose.yml
# Start the Shelly-Asset-Link
docker-compose up --force-recreate
```

## Configuration: 
Shelly-Asset-Link comes with a standard fallback configuration which is invoked when the scan is started via the UI or when started via API call without filters and options.

We use a mounted configuration file instead of hardcoding or environment variables because it allows runtime updates without restarting the container. This is critical for flexibility in production environments where quick adjustments to discovery parameters are needed. By mounting the file, you can change settings on the fly, and the service can reload them automatically.
- Directory: /config inside the container
- File: shelly_al_config.json stores discovery options, filters, and targets
- Mounting in docker-compose.yml

### shelly_al_config.json
```json
{
  "subnet": "192.168.178",
  "startIP": 1,
  "endIP": 254,
  "timeout": 120000,
  "maxParallel": 20
}
```
 - **subnet:** Base network segment for scanning, without trailing dot!
 - **startIP / endIP:** Defines the IP range within the subnet.
 - **timeout:** total scan timeout in milliseconds, once timeout threshold has been hit, 'ctx.Done' will be send, resulting in cancellation of discovery job.
 - **maxParallel:** controls concurrency of scan tasks for faster scanning without overloading asset gateway host.

### Precedence Rules
API request parameters always override the configuration file for the current scan.
If the API request does not provide values, or the discovery is ionvoked via the Industrial Asset Hub UI, the service falls back to `config/shelly_al_config.json`.
Only if the file is missing or invalid, built-in defaults (`scanner/shellyscanner.go`) are used.

## Usage: 
Either start the discovery over the Industrial Asset Hub UI or via API Call.
### 1. Using Industrial Asset Hub UI
Go to the Inbox, start a scan, select the asset-gateway to perform the scan on. If shelly-asset-link has been successfully registered with the gateway, you can select it as scanner and start the scan. A modal will appear and inform you about the status of the discovery. You can as well check the jobs list and find all relevant information for this scan job. The scan will be performed using the preconfigured filters and options. After completion all discovered Shelly devices can be seen in the Inbox of Inustrial Asset Hub.
### 2. Using API calls
to start a scan job via API call, you need to consider some delimiters.
1. You need to set up a server user in your [Siemens Xcelerator Admin Console](https://cloud.sw.siemens.com/admin)
3. You need to generate an access token for your server user
   ```bash
   export TOKEN="$(\
     curl -s -X POST "https://iahprod.eu1.sws.siemens.com/oauth/token" \
       -H "Content-Type: application/x-www-form-urlencoded" \
       -d "grant_type=client_credentials&client_id=<client_id>&client_secret=<client_secret>" \
     | jq -r '.access_token')"
   ```
5. Pre-built example script to call the Industrial Asset Hub API
   ```bash
    #!/bin/bash
    # SPDX-FileCopyrightText: 2025 Michael Leipold github.com/mlp0911
    # SPDX-License-Identifier: MIT
    
    # --- Configuration ---
    IAH_BASE="https://cloud.eu1.sws.siemens.com/api/assethubapi"
    GATEWAY_ID="${GATEWAY_ID:-<gateway-id>}"
    ASSET_LINK_ID="${ASSET_LINK_ID:-cdm-device-class-driver-mlp0911.cdm.al.shelly}"
    DEVICE_CLASS="${DEVICE_CLASS:-mlp0911.cdm.al.shelly}"
    TOKEN="${TOKEN:?set TOKEN as environment variable}"
    
    # --- Prepare payload ---
    # --- set ipRange below ---
    # --- options timeout and maxParallel are currently not considered ---
    # --- these can be set in the shelly_al_config.json ---
   
    PAYLOAD_JSON=$(cat <<EOF
    {
      "assetLinkId": "$ASSET_LINK_ID",
      "deviceClass": "$DEVICE_CLASS",
      "scanMode": "FULL",
      "jobType": "DISCOVERY",
      "discoveryFilters": [
        {"key":"IPRange","comparisonOperator":"EQUAL","value":"192.168.178.1-192.168.178.254"}
      ],
      "discoveryOptions": [
        {"key":"timeoutMs","value":"300000"},
        {"key":"maxParallel","value":"10"}
      ]
    }
    EOF
    )
    
    # --- Start discovery-job ---
    RESP_FILE="/tmp/iah_post.json"
    echo "starting discovery-job for gateway $GATEWAY_ID ..."
    curl -sS -k --http1.1 \
      -H "Authorization: Bearer $TOKEN" \
      -H "Accept: application/json" \
      -H "Content-Type: application/json" \
      --data-binary "$PAYLOAD_JSON" \
      "$IAH_BASE/v1/discovery/v1-earlyaccess/gateways/$GATEWAY_ID/discoveries" \
      --write-out "\nHTTP_CODE:%{http_code}\n" \
    | tee "$RESP_FILE"
    
    # --- extract HTTP-code und job-ID ---
    HTTP_CODE=$(sed -n 's/^HTTP_CODE:\([0-9]\+\)$/\1/p' "$RESP_FILE")
    JOB_ID=$(jq -r '.jobId // .id // .data.id // empty' "$RESP_FILE")
    
    echo "HTTP_CODE=$HTTP_CODE"
    echo "JOB_ID=$JOB_ID"
    
    # --- check job-ID ---
    if [ -z "$JOB_ID" ]; then
      echo "No jobId found in response – cancelled."
      exit 1
    fi
    
    echo "Discovery-Job started successfully. Please use JOB_ID=$JOB_ID for status-polling oder cancel."
    echo "Example for status-polling:"
    echo "curl -sS -k -H \"Authorization: Bearer \$TOKEN\" -H \"Accept: application/json\" \\"
    echo "  \"$IAH_BASE/v1/discovery/v1-earlyaccess/discoveries/$JOB_ID\" | jq ."
   ```

   For further information, how to read the Industrial Asset Hub Inbox, onboard devices from inbox to asset list and alike, please refer to the [Industrial Asset Hub Documentation](https://industrial-assets.io).


## Features: 
Highlight integration capabilities and supported device types. 
## Contribution & License: 
Guidelines for contributing and license details.

## Supported Shelly Devices for RPC Endpoints 
All Shelly devices running the new RPC firmware (Gen2/Gen3, ESP32‑based) provide valid JSON responses for both `.../rpc/Shelly.GetDeviceInfo` and `.../rpc/GetStatus`. These endpoints allow you to retrieve device metadata and current operational status in a consistent way and are being used by `scanner/shellyscanner.go`.

### Supported Devices (as of December 2025)

- **Shelly Plus Series (Gen2)**
  - Plus 1 / 1PM
  - Plus 2PM
  - Plus i4 / i4DC
  - Plus Plug S
  - Plus HT
  - Plus Smoke
  - Plus Add‑On

- **Shelly Pro Series (Gen2, DIN‑Rail)**
  - Pro 1 / 1PM
  - Pro 2 / 2PM
  - Pro 3 / 3EM
  - Pro 4PM

- **Shelly Gen3 Series**
  - 1 Gen3 / 1PM Gen3
  - 2PM Gen3
  - Plug Gen3
  - Dimmer Gen3
  - RGBW Gen3

- **Special Devices with RPC Firmware**
  - Shelly EM (newer revisions)
  - Shelly BLU sensors (via RPC bridge)
  - Shelly Wave (via RPC gateway)

### Not Supported Devices
All **Gen1 devices** (e.g., Shelly 1, 2.5, Dimmer 2, RGBW2, older Plug S) only use the legacy REST API and do not respond to the required RPC endpoints.

### Good Read and More Infos on Shelly IoT Automation Devices
- Official website of [Shelly Group SE](https://www.shelly.com/pages/location-selector).
- Overview of [Shelly Portfolio](https://www.shelly.com/collections/all-products).
- Shelly Developers [API Documentation](https://shelly-api-docs.shelly.cloud/). 
  The scanner logic builds on these well-documented and openly available API calls, e.g. [`.../rpc/Shelly.GetStatus`](https://shelly-api-docs.shelly.cloud/gen2/ComponentsAndServices/Shelly#shellygetstatus-example)
