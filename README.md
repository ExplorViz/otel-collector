# ExplorViz OTel Collector

The ExplorViz OTel Collector is a custom distribution of the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/). It receives incoming telemetry data from OpenTelemetry-compliant agents and interprets it for the purpose of visualization using custom components. The processed telemetry data is exported using a pre-defined set of exporters, including a custom Kafka-based exporter for communication with the [landscape-service](https://github.com/ExplorViz/landscape-service/).

For development instructions, continue reading below. If you just want to run ExplorViz locally, refer to our [Deployment repository](https://github.com/ExplorViz/deployment/) instead.

## Development instructions

### Prerequisites

- [Go](https://go.dev/), version 1.25.0 or higher
- [Protobuf Compiler](https://protobuf.dev/installation/), version 3.0 or higher
- A code editor, such as [Visual Studio Code](https://code.visualstudio.com/)
- Make sure to run the [ExplorViz software stack](../deployment) before starting the service, as it provides the required database(s) and the Kafka broker

### Running the OpenTelemetry Collector Builder

The custom OpenTelemetry Collector binary is built using the [OpenTelemetry Collector Builder](https://opentelemetry.io/docs/collector/extend/ocb/) (`ocb`). It creates a customized binary using the configuration provided in the `builder-config.yaml`. The builder must be re-run for every modification to either the components or to the builder configuration. We include `ocb` as a [tool dependency](https://go.dev/doc/modules/managing-dependencies#tools). To build the project, use:

```shell
go tool builder --config builder-config.yaml
```

The resulting files and the binary will be located in `./cmd/explorviz-otelcol`. To then run the project:

```console
$ cd cmd/explorviz-otelcol
$ go run . --config ../../collector-config-default.yaml
```

Note that you need to run the builder every time before running to see your changes. For a simplified workflow, we recommend [using the provided Makefile](#using-make).

### Code Generation

This project relies on `go:generate` directives to automatically generate code. Code generation can be triggered manually using:

```shell
go generate ./...
```

Make sure to run the above command when updating any `.proto` files to update the corresponding Go bindings. Additionally, the OpenTelemetry Collector uses a `metadata.yaml` file for each component to generate documentation and metadata used by the component's factory. Whenever updating such a file, use `go generate` to run the [mdatagen](https://github.com/open-telemetry/opentelemetry-collector/blob/main/cmd/mdatagen/README.md) tool, which is included as a tool dependency.

### Using Make

For convenience, we provide a Makefile to run common commands from the repository root:

```console
$ make build # Generate and build the project using the default builder config

$ make run   # Generate, build and run in a single step

$ make lint  # Run golangci-lint (must be installed locally)

$ make test  # Generate and run tests

$ make       # Generate, build, lint and test
```

### Code style

As part of our CI/CD pipeline, your code is linted and checked for formatting using [golangci-lint](https://github.com/golangci/golangci-lint), which you can also install locally to lint your code yourself prior to pushing. We recommend using the [official Visual Studio Code extension for Go](https://marketplace.visualstudio.com/items?itemName=golang.go) as well as [configuring the extension for golangci-lint](https://golangci-lint.run/docs/welcome/integrations/) to detect and fix linting / formatting issues as you're working.
