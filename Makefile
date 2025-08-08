.PHONY: build test clean run-node run-dashboard build-cli help

all: build

build: build-cli
	go mod tidy
	go build -o bin/fringe-edge cmd/edge/main.go
	go build -o bin/fringe-dashboard cmd/dashboard/main.go

build-cli:
	cd cli && cargo build --release
	cp cli/target/release/fringe-cli bin/

test:
	go test ./tests/...
	cd cli && cargo test

clean:
	rm -rf bin/
	cd cli && cargo clean

run-node:
	go run cmd/edge/main.go --bootstrap --port 8080 --metrics-port 9090

run-node-join:
	go run cmd/edge/main.go --node 127.0.0.1:8080 --port 8081 --metrics-port 9091

run-dashboard:
	go run cmd/dashboard/main.go 8080

bin:
	mkdir -p bin
