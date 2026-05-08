#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2025 Machine Builder AG
#
# SPDX-License-Identifier: MIT

nohup go run -tags webserver main.go &
bash ./testscripts/wait_till_al_is_started.sh

ASSET_ENDPOINT_PORT=${ASSET_ENDPOINT_PORT:-localhost:8081}
SHELLY_IP_RANGE=${SHELLY_IP_RANGE:-127.0.0.1-127.0.0.1}

# Build discovery file with base64-encoded IPRange filter
DISCOVERY_FILE=$(mktemp /tmp/shelly-discovery-XXXXXX.json)
IP_RANGE_B64=$(printf '%s' "${SHELLY_IP_RANGE}" | base64 | tr -d '\n')

cat > "${DISCOVERY_FILE}" <<EOF
{
  "filters": [
    {
      "key": "IPRange",
      "value": { "raw_data": "${IP_RANGE_B64}" }
    }
  ]
}
EOF

# Run the tests
./al-ctl test api -e ${ASSET_ENDPOINT_PORT} --service-name discovery --discovery-file "${DISCOVERY_FILE}"

rm -f "${DISCOVERY_FILE}"
