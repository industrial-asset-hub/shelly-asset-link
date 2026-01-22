#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2025 Michael Leipold (hhtps://github.com/mlp0911)
#
# SPDX-License-Identifier: MIT

nohup go run -tags webserver main.go &
bash ./testscripts/wait_till_al_is_started.sh

ASSET_ENDPOINT_PORT=${ASSET_ENDPOINT_PORT:-localhost:8081}

# Discover assets on the specified endpoint
./al-ctl assets discover -e ${ASSET_ENDPOINT_PORT}  
