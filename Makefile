.PHONY: all clean build deps run help dist test test-short test-coverage

all: build

# Create dist directory
dist:
	mkdir -p dist
	touch .env

build: dist
	mkdir -p dist/bin
	go build -o dist/bin/certchecker cmd/certchecker/main.go
	cp .env dist/bin/

clean:
	rm -rf dist
	rm -f .env

# Cross-compile for multiple platforms
build-all: dist
	mkdir -p dist/bin/linux-amd64 dist/bin/linux-arm64 \
		dist/bin/darwin-amd64 dist/bin/darwin-arm64 \
		dist/bin/windows-amd64 dist/bin/windows-x86 \
		dist/bin/windows-arm64 dist/bin/windows-arm32
	# Linux (amd64)
	GOOS=linux GOARCH=amd64 go build -o dist/bin/linux-amd64/certchecker cmd/certchecker/main.go
	cp .env dist/bin/linux-amd64/
	# Linux (arm64)
	GOOS=linux GOARCH=arm64 go build -o dist/bin/linux-arm64/certchecker cmd/certchecker/main.go
	cp .env dist/bin/linux-arm64/
	# macOS (amd64)
	GOOS=darwin GOARCH=amd64 go build -o dist/bin/darwin-amd64/certchecker cmd/certchecker/main.go
	cp .env dist/bin/darwin-amd64/
	# macOS (arm64/M1)
	GOOS=darwin GOARCH=arm64 go build -o dist/bin/darwin-arm64/certchecker cmd/certchecker/main.go
	cp .env dist/bin/darwin-arm64/
	# Windows (amd64)
	GOOS=windows GOARCH=amd64 go build -o dist/bin/windows-amd64/certchecker.exe cmd/certchecker/main.go
	cp .env dist/bin/windows-amd64/
	# Windows (x86/32-bit)
	GOOS=windows GOARCH=386 go build -o dist/bin/windows-x86/certchecker.exe cmd/certchecker/main.go
	cp .env dist/bin/windows-x86/
	# Windows (ARM64)
	GOOS=windows GOARCH=arm64 go build -o dist/bin/windows-arm64/certchecker.exe cmd/certchecker/main.go
	cp .env dist/bin/windows-arm64/
	# Windows (ARM32)
	GOOS=windows GOARCH=arm go build -o dist/bin/windows-arm32/certchecker.exe cmd/certchecker/main.go
	cp .env dist/bin/windows-arm32/

# Run targets
run:
	cd dist/bin && ./certchecker

# Install dependencies
deps:
	go mod tidy

# Test targets
test:
	go test ./...

test-short:
	go test -short ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Distribution packages
dist-package: clean dist build
	cd dist && tar czf certchecker.tar.gz bin
	@echo "Distribution package created at dist/certchecker.tar.gz"

# Cross-platform distribution package
dist-package-all: clean dist build-all
	cd dist && tar czf certchecker-linux-amd64.tar.gz bin/linux-amd64/*
	cd dist && tar czf certchecker-linux-arm64.tar.gz bin/linux-arm64/*
	cd dist && tar czf certchecker-darwin-amd64.tar.gz bin/darwin-amd64/*
	cd dist && tar czf certchecker-darwin-arm64.tar.gz bin/darwin-arm64/*
	cd dist && zip -j certchecker-windows-amd64.zip bin/windows-amd64/*
	cd dist && zip -j certchecker-windows-x86.zip bin/windows-x86/*
	cd dist && zip -j certchecker-windows-arm64.zip bin/windows-arm64/*
	cd dist && zip -j certchecker-windows-arm32.zip bin/windows-arm32/*
	@echo "Cross-platform distribution packages created in dist/"

# Help
help:
	@echo "Available targets:"
	@echo "  make build            - Build for current platform"
	@echo "  make build-all        - Build for all platforms"
	@echo "  make run              - Run from dist"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make deps             - Install dependencies"
	@echo "  make test             - Run all tests"
	@echo "  make test-short       - Run tests excluding integration tests"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make dist-package     - Create distributable package for current platform"
	@echo "  make dist-package-all - Create distributable packages for all platforms" 