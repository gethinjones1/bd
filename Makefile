.PHONY: test build test-install
test:
	go test ./...
build:
	go build -o bin/bd ./cmd/bd
test-install:
	sh scripts/install_test.sh
