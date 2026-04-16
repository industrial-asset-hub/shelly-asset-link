/*
 * SPDX-FileCopyrightText: 2025 Michael Leipold (https://github.com/mlp0911)
 *
 * SPDX-License-Identifier: MIT
 *
 * Author: Michael Leipold <https://github.com/mlp0911>
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"shelly-asset-link/handler"

	"github.com/industrial-asset-hub/asset-link-sdk/v3/assetlink"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/logging"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/metadata"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// values provided by linker
	version = "1.1.1"
	commit  = "unknown"
	date    = "unknown"

	// values provided by cookiecutter
	alId   = "mlp0911.cdm.al.shelly"
	alName = "Shelly Asset Link"
	vendor = "Michael Leipold"
)

func main() {
	logging.SetupLogging()
	log.Info().
		Str("Version", version).
		Str("commit", commit).
		Str("date", date).
		Msg("Starting " + alName + " driver")

	// Setup log of log infrastructure
	var logLevel string
	var grpcServerAddress, grpcServerEndpointAddress, httpServerAddress, registryAddress string
	flag.StringVar(&logLevel, "log-level", "info", fmt.Sprintf("set log level. one of: %s,%s,%s,%s,%s,%s,%s",
		zerolog.TraceLevel.String(),
		zerolog.DebugLevel.String(),
		zerolog.InfoLevel.String(),
		zerolog.WarnLevel.String(),
		zerolog.ErrorLevel.String(),
		zerolog.FatalLevel.String(),
		zerolog.PanicLevel.String()))
	flag.StringVar(&grpcServerAddress, "grpc-server-address", "localhost:8081", "gRPC server endpoint")
	flag.StringVar(&grpcServerEndpointAddress, "grpc-server-endpoint-address", "localhost", "Address which is registered")
	flag.StringVar(&httpServerAddress, "http-address", "localhost:8082", "HTTP server endpoint")
	flag.StringVar(&registryAddress, "grpc-registry-address", "grpc-server-registry:50051", "gRPC registry address")

	// Parse the CLI flags
	flag.Parse()

	// Set log level
	logging.AdjustLogLevel(logLevel)

	// Register AL implementation
	alImpl := new(handler.AssetLinkImplementation)
	alInstance := assetlink.New(metadata.Metadata{
		Version: metadata.Version{Version: version, Commit: commit, Date: date},
		AlId:    alId,
		AlName:  alName,
		Vendor:  vendor,
	}).
		Discovery(alImpl).
		Build()

	// Signal handler for a proper shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(d *assetlink.AssetLink) {
		<-c
		log.Info().Msg("Received SIGTERM. Exiting.")
		d.Stop()
		os.Exit(1)
	}(alInstance)

	// Start asset link
	if err := alInstance.Start(grpcServerAddress, grpcServerEndpointAddress, registryAddress, httpServerAddress); err != nil {
		log.Fatal().Err(err).Msg("Could not start asset link instance")
	}

}
