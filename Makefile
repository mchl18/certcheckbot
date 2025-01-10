.PHONY: all clean build-go deps-go run-go help dist

all: build-go

# Create dist directory
dist:
	mkdir -p dist
	touch .env

build-go: dist
	mkdir -p dist/go
	cd go && go build -o ../dist/go/certchecker cmd/certchecker/main.go
	cp .env dist/go/

clean:
	rm -rf dist
	rm -f .env

# Cross-compile Go for multiple platforms
build-go-all: dist
	# Linux (amd64)
	cd go && GOOS=linux GOARCH=amd64 go build -o ../dist/go/certchecker-linux-amd64 cmd/certchecker/main.go
	# Linux (arm64)
	cd go && GOOS=linux GOARCH=arm64 go build -o ../dist/go/certchecker-linux-arm64 cmd/certchecker/main.go
	# macOS (amd64)
	cd go && GOOS=darwin GOARCH=amd64 go build -o ../dist/go/certchecker-darwin-amd64 cmd/certchecker/main.go
	# macOS (arm64/M1)
	cd go && GOOS=darwin GOARCH=arm64 go build -o ../dist/go/certchecker-darwin-arm64 cmd/certchecker/main.go
	# Windows (amd64)
	cd go && GOOS=windows GOARCH=amd64 go build -o ../dist/go/certchecker-windows-amd64.exe cmd/certchecker/main.go
	# Windows (x86/32-bit)
	cd go && GOOS=windows GOARCH=386 go build -o ../dist/go/certchecker-windows-x86.exe cmd/certchecker/main.go
	# Windows (ARM64)
	cd go && GOOS=windows GOARCH=arm64 go build -o ../dist/go/certchecker-windows-arm64.exe cmd/certchecker/main.go
	# Windows (ARM32)
	cd go && GOOS=windows GOARCH=arm go build -o ../dist/go/certchecker-windows-arm32.exe cmd/certchecker/main.go
	cp .env dist/go/

# Run targets
run-go:
	cd dist/go && ./certchecker

# Install dependencies
deps-go:
	cd go && go mod tidy

# Distribution packages
dist-package: clean dist build-go
	cd dist && tar czf certchecker.tar.gz go
	@echo "Distribution package created at dist/certchecker.tar.gz"

# Cross-platform distribution package
dist-package-all: clean dist build-go-all
	cd dist && tar czf certchecker-linux-amd64.tar.gz go/certchecker-linux-amd64 go/.env
	cd dist && tar czf certchecker-linux-arm64.tar.gz go/certchecker-linux-arm64 go/.env
	cd dist && tar czf certchecker-darwin-amd64.tar.gz go/certchecker-darwin-amd64 go/.env
	cd dist && tar czf certchecker-darwin-arm64.tar.gz go/certchecker-darwin-arm64 go/.env
	cd dist && zip -r certchecker-windows-amd64.zip go/certchecker-windows-amd64.exe go/.env
	cd dist && zip -r certchecker-windows-x86.zip go/certchecker-windows-x86.exe go/.env
	cd dist && zip -r certchecker-windows-arm64.zip go/certchecker-windows-arm64.exe go/.env
	cd dist && zip -r certchecker-windows-arm32.zip go/certchecker-windows-arm32.exe go/.env
	@echo "Cross-platform distribution packages created in dist/"

# Help
help:
	@echo "Available targets:"
	@echo "  make build-go          - Build Go version for current platform"
	@echo "  make build-go-all      - Build Go version for all platforms"
	@echo "  make run-go            - Run Go version from dist"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make deps-go           - Install Go dependencies"
	@echo "  make dist-package      - Create distributable package for current platform"
	@echo "  make dist-package-all  - Create distributable packages for all platforms" 