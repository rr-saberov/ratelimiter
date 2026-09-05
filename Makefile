BINARY_NAME := ratelimiter
MAIN_PACKAGE := ./cmd/ratelimiter
VERSION ?= dev

.PHONY: build test lint run clean docker

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY_NAME) $(MAIN_PACKAGE)

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME) coverage.out
	rm -rf dist

docker:
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY_NAME):$(VERSION) .
