#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2025 Michael Leipold (hhtps://github.com/mlp0911)
#
# SPDX-License-Identifier: MIT

nohup go run -tags webserver main.go &
bash ./testscripts/wait_till_al_is_started.sh

ASSET_ENDPOINT_PORT=${ASSET_ENDPOINT_PORT:-localhost:8081}

# echo "OS_NAME: ${OS_NAME}"
# echo "ARCH_NAME: ${ARCH_NAME}"
export OS_NAME=linux
export ARCH_NAME=amd64
# curl -L -o al-ctl_${OS_NAME}_${ARCH_NAME}.tar.gz https://github.com/industrial-asset-hub/asset-link-sdk/releases/download/v3.4.3/al-ctl_${OS_NAME}_${ARCH_NAME}.tar.gz
curl -L -o al-ctl_Linux_x86_64.tar.gz https://github.com/industrial-asset-hub/asset-link-sdk/releases/download/v3.6.2/al-ctl_Linux_x86_64.tar.gz
# tar -xf al-ctl_${OS_NAME}_${ARCH_NAME}.tar.gz
tar -xf al-ctl_Linux_x86_64.tar.gz
chmod +x al-ctl

# Discover assets on the specified endpoint
./al-ctl assets discover -e ${ASSET_ENDPOINT_PORT}  
