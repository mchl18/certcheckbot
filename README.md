# SSL Certificate Checker

A service that monitors SSL certificates for a list of domains and sends alerts to Slack when certificates are nearing expiration.

## Features

- Monitor multiple domains for SSL certificate expiration
- Configurable alert thresholds (e.g., alert at 30, 14, and 7 days before expiration)
- Slack notifications for expiring certificates
- Optional heartbeat messages to confirm service is running
- HTTP API for health checks and log access
- Automatic backup of alert history
- Structured logging with file and console output

## Installation

### Linux/macOS

```bash
# Download and install
curl -sSL https://raw.githubusercontent.com/mchl18/ssl-expiration-check-bot/main/install/install.sh | bash

# Add to your shell configuration (~/.bashrc, ~/.zshrc, etc.):
export PATH="$HOME/.certchecker/bin:$PATH"
```

### Windows

```powershell
# Download and run installer (in PowerShell)
Invoke-Expression (New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/mchl18/ssl-expiration-check-bot/main/install/install.ps1')

# Add to PATH (run in PowerShell as Administrator)
[Environment]::SetEnvironmentVariable('Path', $env:Path + ';$env:USERPROFILE\.certchecker\bin', 'User')
```

After installation, restart your terminal/PowerShell for the PATH changes to take effect.

## Configuration

Run the configuration command:
```bash
certchecker -configure
```

You'll be prompted for:
- Domains to monitor (comma-separated)
- Alert threshold days (comma-separated)
- Slack webhook URL for notifications
- Optional: Heartbeat interval in hours
- Optional: Check interval in hours (defaults to 6)
- Optional: HTTP server settings (enabled/disabled, port, auth token)

Configuration is stored in `$HOME/.certchecker/config/.env`. You can modify this file while the service is stopped without rebuilding.

Example configuration:
```env
# Domains to monitor
DOMAINS=example.com,test.com

# Alert thresholds in days
THRESHOLD_DAYS=7,14,30

# Slack webhook URL for notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx

# Optional: Send heartbeat messages every N hours
HEARTBEAT_HOURS=24

# Optional: Check certificates every N hours (default: 6)
CHECK_INTERVAL_HOURS=6

# Optional: HTTP server settings
HTTP_ENABLED=true
HTTP_PORT=8080
HTTP_AUTH_TOKEN=your-secret-token
```

## Usage

Run the service:
```bash
certchecker
```

The service will:
1. Check certificates for all configured domains
2. Send alerts to Slack if any certificates are expiring soon
3. Send heartbeat messages if configured
4. Start HTTP server if enabled

## HTTP API

When HTTP server is enabled, the following endpoints are available:

### Health Check
```
GET /health
Authorization: Bearer your-secret-token
```

Response:
```json
{
  "status": "ok",
  "uptime": "1h30m45s",
  "domains": ["example.com", "test.com"],
  "thresholds": [7, 14, 30],
  "started_at": "2024-01-13T20:00:00Z",
  "checked_at": "2024-01-13T21:00:00Z",
  "version": "1.0.0"
}
```

### Logs
```
GET /logs?lines=100
Authorization: Bearer your-secret-token
```

Response:
```json
{
  "lines": 100,
  "total": 150,
  "logs": [
    "[2024-01-13T20:00:00Z] [INFO] Starting certificate check",
    "[2024-01-13T20:00:01Z] [INFO] Certificate expiration check for example.com"
  ],
  "timestamp": "2024-01-13T21:00:00Z"
}
```

## Directory Structure

All application data is stored in `$HOME/.certchecker/`:
```
.certchecker/
├── bin/           # Binary files
├── config/        # Configuration files
│   └── .env      # Environment variables
├── logs/          # Log files
│   └── cert-checker.log
└── data/          # Application data
    └── alert-history.json
```

## Development

Clone and build:
```bash
git clone https://github.com/mchl18/certchecker.git
cd certchecker
make test    # Run tests
make build   # Build binary
make install # Install locally
```

## License

MIT License - see [LICENSE](LICENSE) for details. 