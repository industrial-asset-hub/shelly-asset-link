#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
# SPDX-License-Identifier: MIT

set -xeu

systemctl daemon-reload
# shellcheck disable=SC1083
systemctl enable shelly-asset-link.service
# shellcheck disable=SC1083
systemctl start shelly-asset-link.service
