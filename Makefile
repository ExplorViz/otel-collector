.PHONY: all
all: generate build lint test

.PHONY: generate
generate:
	go generate ./...

.PHONY: build
build: generate
	go tool builder --config builder-config.yaml

.PHONY: run
run: build
	cd cmd/explorviz-otelcol && go run . --config ../../collector-config-default.yaml

.PHONY: lint
lint: generate
	golangci-lint run

.PHONY: test
test: generate
	go test ./...