#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
#
# SPDX-License-Identifier: MIT

set -xeu
# shellcheck disable=SC1083
systemctl stop shelly-asset-link

systemctl daemon-reload
