# SSL Certificate Checker

A service that monitors SSL certificates for specified domains and sends alerts via Slack when certificates are approaching expiration. Available in both Go and Node.js implementations.

## Quick Install

```bash
# Install the latest version
curl -sSL https://raw.githubusercontent.com/madbook/certchecker/main/install.sh | bash

# Add to your PATH (add to .bashrc/.zshrc for persistence)
export PATH="$PATH:$HOME/.certchecker/bin"

# Run the interactive configuration
certchecker config
```

## Features

- Monitor multiple domains for SSL certificate expiration
- Configurable alert thresholds (e.g., alert at 45, 30, 14, and 7 days before expiration)
- Slack notifications for expiring certificates
- Detailed logging with process information
- Alert history tracking with automatic backups
- Runs as a service with configurable check intervals
- Runtime configuration via `.env` file (no recompilation needed)
- Bundled distributions for both Go and Node.js versions

## Prerequisites

- For Go version:
  - Go 1.21 or later
- For Node.js version:
  - Node.js 18 or later
  - Yarn package manager
- A Slack webhook URL for notifications

## Configuration

You can configure the certificate checker in two ways:

### 1. Interactive Configuration

Run the interactive configuration wizard:
```bash
certchecker config
```

This will guide you through setting up:
- Domains to monitor
- Alert threshold days
- Slack webhook URL

The configuration will be saved to `$HOME/.certchecker/config/.env`.

### 2. Manual Configuration

Create or edit `.env` file in one of these locations (in order of precedence):
1. `$HOME/.certchecker/config/.env`
2. `.env` in the current directory
3. `.env` in the parent directory

```env
# Comma-separated list of domains to monitor
DOMAINS=example.com,subdomain.example.com

# Comma-separated list of days before expiration to send alerts
THRESHOLD_DAYS=7,14,30,45

# Slack webhook URL for notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/your/webhook/url
```

### Environment Configuration Notes

- The `.env` file is read at runtime, not during compilation
- You can modify the `.env` file while the service is stopped without rebuilding
- The `.env` file is automatically copied to the dist directory during build
- Both versions in the dist directory expect the `.env` file to be in their respective directories

## Installation

The project uses a Makefile for consistent building and running of both versions. All builds are output to the `dist` directory.

### Install Dependencies

For Go version:
```bash
make deps-go
```

For Node.js version:
```bash
make deps-node
```

### Build

Build both versions:
```bash
make all
```

Or build individually:
```bash
make build-go    # Outputs to dist/go/certchecker
make build-node  # Outputs to dist/node/certchecker.js
```

### Distribution Package

Create a distributable package containing both versions:
```bash
make dist-package
```
This creates `dist/certchecker.tar.gz` containing both the Go and Node.js versions with their respective `.env` files.

## Usage

### Running the Service

Run the Go version:
```bash
make run-go     # Runs from dist/go/certchecker
```

Run the Node.js version:
```bash
make run-node   # Runs from dist/node/certchecker.js
```

### Clean Up

Remove build artifacts and dependencies:
```bash
make clean
```

### Available Make Commands

View all available commands:
```bash
make help
```

## Project Structure

```
.
├── .env                # Environment configuration template
├── Makefile           # Build and run commands
├── go/                # Go source code
│   ├── cmd/
│   │   └── certchecker/
│   │       └── main.go
│   └── internal/
│       ├── alert/     # Slack notification
│       ├── checker/   # Certificate checking
│       ├── logger/    # Logging
│       └── storage/   # History management
├── node/              # Node.js source code
│   └── src/
│       ├── alert/     # Slack notification
│       ├── checker/   # Certificate checking
│       ├── logger/    # Logging
│       └── storage/   # History management
├── dist/              # Built artifacts
│   ├── go/           # Go binary and config
│   │   ├── certchecker
│   │   └── .env
│   └── node/         # Node.js bundle and config
│       ├── certchecker.js
│       └── .env
└── logs/              # Log files and history data
    ├── cert-checker.log
    └── data/
        ├── alert-history.json
        └── alert-history.json.backup
```

## Logging

Logs are stored in the `logs` directory:
- `logs/cert-checker.log`: Application logs
- `logs/data/alert-history.json`: Alert history
- `logs/data/alert-history.json.backup`: Automatic backup of alert history

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 