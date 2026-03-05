test-integration:
	go test -v -tags=integration ./tests/integration...

test:
	go test -v ./...

# Run golangci-lint on the project
lint:
	golangci-lint run --fix