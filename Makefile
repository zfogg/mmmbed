.PHONY: build test lint fmt fmt-check clean generate

build:
	go build -o bin/mmmbed ./cmd/mmmbed

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w cmd embed internal examples

fmt-check:
	@test -z "$$(gofmt -l cmd embed internal examples)" || (gofmt -l cmd embed internal examples; exit 1)

generate:
	go generate ./examples/...

clean:
	go clean
