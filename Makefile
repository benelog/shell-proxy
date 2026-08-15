BINARY := shell-proxy

.PHONY: build run fmt lint test check ci dist clean

build:
	go build -o $(BINARY) .

# Cross-compile release binaries into dist/ for the platforms published on the
# GitHub release. CGO is disabled so every target is a static binary.
dist:
	rm -rf dist && mkdir -p dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o dist/$(BINARY)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -o dist/$(BINARY)-linux-arm64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o dist/$(BINARY)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o dist/$(BINARY)-darwin-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY)-windows-amd64.exe .

run: build
	./$(BINARY)

fmt:
	goimports -w .

lint:
	golangci-lint run ./...

test:
	go test ./...

check: fmt lint test

ci: lint test

clean:
	rm -rf $(BINARY) dist
