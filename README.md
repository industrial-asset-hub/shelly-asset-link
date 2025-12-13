# Shelly Asset Link
**shelly-asset-link** is an integration module based on the https://github.com/industrial-asset-hub/asset-link SDK of the Siemens Industrial Asset Hub. It enables the integration of IoT automation components from Shelly Group SE into the Industrial Asset Hub. This allows Shelly devices to be represented as digital assets within industrial applications and ecosystems, providing following advantages:

- **Interoperability:** The asset link seamlessly connects Shelly IoT automation components with the Siemens Industrial Asset Hub. The integration focuses on instantiated device data, providing a digital nameplate and all relevant descriptive metadata for each Shelly device. This ensures that devices are represented consistently as standardized digital assets within the Industrial Asset Hub ecosystem, enabling the use of Shelly devices in digital twins, asset management, and automation processes.

- **Standardized Interfaces:** The asset link provides plug-and-play integration into the industrial-grade inventory service, Industrial Asset Hub, without the need for custom adapters or proprietary protocols.
Built on a future-proof architecture, it enables seamless extension to additional management systems and services, ensuring scalability and long-term interoperability.

## Installation: 
How to set up shelly-asset-link (dependencies, SDK link, Shelly device requirements). 
## Usage: 
Example commands or code snippets for connecting Shelly devices to the Industrial Asset Hub. 
## Configuration: 
Environment variables or configuration files needed. 
## Features: 
Highlight integration capabilities and supported device types. 
## Contribution & License: 
Guidelines for contributing and license details.








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

