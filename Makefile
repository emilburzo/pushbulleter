.PHONY: build install clean test run daemon

# Build the application
build:
	go build -o pushbullet-client cmd/pushbullet-client/main.go

# Install to /usr/local/bin
install: build
	sudo cp pushbullet-client /usr/local/bin/

# Clean build artifacts
clean:
	rm -f pushbullet-client

# Run tests
test:
	go test ./...

# Run in GUI mode
run: build
	./pushbullet-client

# Run in daemon mode
daemon: build
	./pushbullet-client -daemon

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build for different architectures
build-all:
	GOOS=linux GOARCH=amd64 go build -o pushbullet-client-linux-amd64 cmd/pushbullet-client/main.go
	GOOS=linux GOARCH=arm64 go build -o pushbullet-client-linux-arm64 cmd/pushbullet-client/main.go

# Create release package
package: build-all
	mkdir -p dist
	cp pushbullet-client-linux-amd64 dist/
	cp pushbullet-client-linux-arm64 dist/
	cp README.md dist/
	tar -czf dist/pushbullet-client-linux.tar.gz -C dist .
