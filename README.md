# Shelly Asset Link
**shelly-asset-link** is an integration module based on the https://github.com/industrial-asset-hub/asset-link SDK of the Siemens Industrial Asset Hub. It enables the integration of IoT automation components from Shelly Group SE into the Industrial Asset Hub. This allows Shelly devices to be represented as digital assets within industrial applications and ecosystems, providing following advantages:

- **Interoperability:** Seamlessly connects Shelly IoT automation components with the Siemens Industrial Asset Hub.
- **Standardized Interface:** Leverages the Asset-Link SDK for consistent data integration.
- **Industry 4.0 Value:** Enables the use of Shelly devices in digital twins, asset management, and automation processes.
- **Extensibility:** Provides a foundation for integrating additional IoT devices and services.












This is a automatically created IAH Asset Link project.

Before starting, you should synchronize the Go modules. This can be done by:

```bash
$ go mod tidy
[...]
```

## Run && Building

To start the AssetLink execute:

```bash
# Execute
$ go run -tags webserver main.go --grpc-server-address=$(hostname -i):8080 \
--grpc-server-endpoint-address=$(hostname -i) --grpc-registry-address=localhost:50051 \
--http-address=$(hostname -i):8081
[...]
```

To create a release:

```bash
$ goreleaser release --snapshot --clean
$ ls -al dist/
[...]
```

## Run the Pipeline
The reference Asset Link includes GitHub workflow that automates the pipeline steps.
The pipeline will be executed automatically during push operation or creation of pull request.

