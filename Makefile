.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint: fmt
	golangci-lint run

.PHONY: test
test:
	go test ./... -v -cover -race