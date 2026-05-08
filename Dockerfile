# SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
#
# SPDX-License-Identifier: MIT

FROM golang:alpine AS builder
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -tags webserver -trimpath \
    -ldflags="-s -w -extldflags '-static' \
              -X main.version=${VERSION} \
              -X main.commit=${COMMIT} \
              -X main.date=${DATE}" \
    -o /al .

FROM scratch
COPY --from=builder --chmod=0755 /al /al
EXPOSE 8081
ENTRYPOINT ["/al"]
