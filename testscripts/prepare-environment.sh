#!/usr/bin/env bash

# SPDX-FileCopyrightText: Michael Leipold (https://github.com/mlp0911)
#
# SPDX-License-Identifier: MIT

# Detect OS and ARCH if not already set
if [ -z "${OS_NAME}" ]; then
    export OS_NAME=$(uname -s)
fi

if [ -z "${ARCH_NAME}" ]; then
    export ARCH_NAME=$(uname -m)
fi

echo "OS_NAME: ${OS_NAME}"
echo "ARCH_NAME: ${ARCH_NAME}"

# Download and prepare al-ctl
curl -L -o al-ctl.tar.gz https://github.com/industrial-asset-hub/asset-link-sdk/releases/download/v3.6.2/al-ctl_${OS_NAME}_${ARCH_NAME}.tar.gz
tar -xf al-ctl.tar.gz
chmod +x al-ctl
