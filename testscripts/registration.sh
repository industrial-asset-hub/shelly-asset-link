#!/usr/bin/env bash

# SPDX-FileCopyrightText: Michael Leipold (https://github.com/mlp0911)
#
# SPDX-License-Identifier: MIT

nohup go run -tags webserver main.go --grpc-registry-address=localhost:50051 &
bash ./testscripts/wait_till_al_is_started.sh

ASSET_ENDPOINT_PORT=${ASSET_ENDPOINT_PORT:-localhost:8081}
GRPC_SERVER_REGISTRY=${GRPC_SERVER_REGISTRY:-localhost:50051}

./al-ctl list

# To validate registration of asset link
./al-ctl test registration -e ${ASSET_ENDPOINT_PORT} -r ${GRPC_SERVER_REGISTRY} -f ./registry.json
