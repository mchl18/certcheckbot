# SSL Certificate Checker

A service that monitors SSL certificates for specified domains and sends alerts via Slack when certificates are approaching expiration.

## Quick Install

```bash
# Install the latest version
curl -sSL https://raw.githubusercontent.com/mchl18/certcheckbot/main/install.sh | bash

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

## Prerequisites

- Go 1.21 or later
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

Create or edit `.env` file at:
`$HOME/.certchecker/config/.env`

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
- All application data is stored in the `$HOME/.certchecker` directory

## Installation

The project uses a Makefile for consistent building and running. All builds are output to the `dist` directory.

### Install Dependencies

```bash
make deps
```

### Build

Build the project:
```bash
make build
```

### Distribution Package

Create a distributable package:
```bash
make dist-package
```
This creates platform-specific packages in the `dist` directory.

## Usage

### Running the Service

```bash
make run
```

### Clean Up

Remove build artifacts:
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
$HOME/.certchecker/
├── config/           # Configuration directory
│   └── .env         # Environment configuration
├── logs/            # Application logs
│   └── cert-checker.log
└── data/            # Application data
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