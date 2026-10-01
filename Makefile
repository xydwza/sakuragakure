.PHONY: build run test generate vet fmt

BINARY := sakuragakure

build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

run:
	go run ./cmd/$(BINARY) serve

vet:
	go vet ./...

test: vet
	go test ./...

generate:
	templ generate
	sqlc generate

fmt:
	gofmt -w .

clean:
	rm -rf bin
