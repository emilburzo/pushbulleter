.PHONY: build install clean test run daemon

# Build the application
build:
	go build -o pushbulleter cmd/pushbulleter/main.go

# Install to /usr/local/bin
install: build
	sudo cp pushbulleter /usr/local/bin/

# Clean build artifacts
clean:
	rm -f pushbulleter

# Run tests
test:
	go test ./...

# Run in GUI mode
run: build
	./pushbulleter


# Download dependencies
deps:
	go mod download
	go mod tidy

# Build for different architectures
build-all:
	GOOS=linux GOARCH=amd64 go build -o pushbulleter-linux-amd64 cmd/pushbulleter/main.go
	GOOS=linux GOARCH=arm64 go build -o pushbulleter-linux-arm64 cmd/pushbulleter/main.go

# Create release package
package: build-all
	mkdir -p dist
	cp pushbulleter-linux-amd64 dist/
	cp pushbulleter-linux-arm64 dist/
	cp README.md dist/
	tar -czf dist/pushbulleter-linux.tar.gz -C dist .
