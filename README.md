# SSL Certificate Checker

A service that monitors SSL certificates for a list of domains and sends alerts to Slack when certificates are nearing expiration.

## Features

- Monitors SSL certificates for multiple domains
- Configurable alert thresholds and Slack notifications for expiring certificates
- Optional heartbeat messages to confirm service is running
- HTTP API for health checks and log retrieval
- All application data stored in `$HOME/.certchecker` directory

## Installation

```bash
# Download and install the latest version
curl -sSL https://raw.githubusercontent.com/mchl18/certchecker/main/install.sh | bash

# Add the installation directory to your PATH
echo 'export PATH="$HOME/.certchecker/bin:$PATH"' >> ~/.zshrc  # or ~/.bashrc
source ~/.zshrc  # or ~/.bashrc

# Run the configuration command
certchecker config
```

## Configuration

Run `certchecker config` to create or update the configuration. You'll be prompted for:

- Domains to monitor (comma-separated)
- Alert thresholds in days (comma-separated)
- Slack webhook URL for notifications
- Optional heartbeat interval in hours
- Optional check interval in hours (default: 6)
- Optional HTTP server settings:
  - Enable/disable HTTP server
  - Server port (default: 8080)
  - Authentication token

The configuration is stored in `$HOME/.certchecker/config/.env`. You can modify this file while the service is stopped without rebuilding.

Example configuration:
```env
# Domains to monitor
DOMAINS=example.com,test.com

# Alert thresholds in days
THRESHOLD_DAYS=7,14,30

# Slack webhook URL for notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx/yyy/zzz

# Optional: Send heartbeat messages every N hours
HEARTBEAT_HOURS=24

# Optional: Check certificates every N hours
INTERVAL_HOURS=6

# Optional: HTTP server settings
HTTP_ENABLED=true
HTTP_PORT=8080
HTTP_AUTH_TOKEN=your-secret-token
```

## HTTP API

When enabled, the HTTP server provides the following endpoints:

### Health Check

```http
GET /health
Authorization: Bearer your-secret-token
```

Response:
```json
{
  "status": "ok",
  "uptime": "3h2m15s",
  "domains": ["example.com", "test.com"],
  "thresholds": [7, 14, 30],
  "started_at": "2024-01-01T12:00:00Z",
  "checked_at": "2024-01-01T14:30:00Z",
  "version": "1.0.0"
}
```

### Logs

```http
GET /logs?lines=100
Authorization: Bearer your-secret-token
```

Response:
```json
{
  "lines": 100,
  "total": 500,
  "logs": [
    "2024-01-01 12:00:00 [INFO] Service started",
    "2024-01-01 12:00:01 [INFO] Checking certificates..."
  ],
  "timestamp": "2024-01-01T15:00:00Z"
}
```

## Directory Structure

```
$HOME/.certchecker/
├── bin/           # Binary files
├── config/        # Configuration files
│   └── .env
├── data/          # Application data
│   └── alert-history.json
└── logs/          # Log files
    └── cert-checker.log
```

## Development

```bash
# Clone the repository
git clone https://github.com/mchl18/certchecker.git
cd certchecker

# Build the project
make build

# Run tests
make test

# Install locally
make install
```

## License

MIT License - see [LICENSE](LICENSE) for details. 