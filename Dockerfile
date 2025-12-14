# SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
#
# SPDX-License-Identifier: MIT

FROM scratch

COPY --chmod=0755 shelly-asset-link /al

EXPOSE 8081

ENTRYPOINT ["/al"]

