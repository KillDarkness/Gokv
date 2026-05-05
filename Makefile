.PHONY: run build test fmt vet clean

run:
	go run ./cmd/gokv

build:
	mkdir -p bin
	go build -o bin/gokv ./cmd/gokv

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf bin dist coverage.out
