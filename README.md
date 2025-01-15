# SSL Certificate Checker

A service that monitors SSL certificates for a list of domains and sends alerts to Slack when certificates are nearing expiration.

## Features

- Monitor multiple domains for SSL certificate expiration
- Configurable alert thresholds (e.g., alert at 30, 14, and 7 days before expiration)
- Slack notifications for expiring certificates
- Optional heartbeat messages to confirm service is running
- HTTP API for health checks and log access
- Web UI for configuration and log viewing
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

## Setting up Slack Webhook

Before configuring the service, you'll need to create a Slack webhook:

1. Go to [Slack Apps](https://api.slack.com/apps) and click "Create New App"
2. Choose "From scratch" and give your app a name (e.g., "SSL Certificate Checker")
3. Select the workspace where you want to receive notifications
4. In the sidebar, under "Features", click "Incoming Webhooks"
5. Toggle "Activate Incoming Webhooks" to On
6. Click "Add New Webhook to Workspace"
7. Choose the channel where you want to receive notifications
8. Copy the Webhook URL - you'll need this during configuration

The webhook URL will look something like: `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`

You can test your webhook with curl:
```bash
curl -X POST -H 'Content-type: application/json' --data '{"text":"Hello from SSL Certificate Checker!"}' YOUR_WEBHOOK_URL
```

## Configuration

You can configure the service in three ways:

### 1. Interactive Setup

Simply run the service without any flags:
```bash
certchecker
```

You'll be prompted to choose between web UI or command line configuration.

### 2. Using the Web UI

Start the service with the web UI:
```bash
certchecker -webui
```

Then open http://localhost:8081 in your browser. On first run, you'll be prompted to configure the service. After configuration, you'll need to use your chosen authentication token to log in.

### 3. Using the Command Line

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

Configuration is stored in `$HOME/.certchecker/config/config.yaml`. You can modify this file while the service is stopped without rebuilding.

Example configuration:
```yaml
# Domains to monitor
domains:
  - example.com
  - test.com

# Alert thresholds in days
threshold_days:
  - 7
  - 14
  - 30

# Slack webhook URL for notifications
slack_webhook_url: https://hooks.slack.com/services/xxx

# Optional: Send heartbeat messages every N hours
heartbeat_hours: 24

# Optional: Check certificates every N hours (default: 6)
interval_hours: 6

# Optional: HTTP server settings
http_enabled: true
http_port: 8080
http_auth_token: your-secret-token
```

## Usage

Run the service:
```bash
# Run with web UI (recommended)
certchecker -webui

# Run without web UI
certchecker
```

The service will:
1. Check certificates for all configured domains
2. Send alerts to Slack if any certificates are expiring soon
3. Send heartbeat messages if configured
4. Start HTTP server if enabled
5. Start web UI if -webui flag is used

## Web UI

The web interface provides:
- Initial setup wizard
- Configuration management
- Log viewing
- Token-based authentication

Access the web UI at http://localhost:8081 after starting with the `-webui` flag.

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
│   └── config.yaml # Configuration file
├── logs/          # Log files
│   └── cert-checker.log
└── data/          # Application data
    └── alert-history.json
```

## Development

Clone and build:
```bash
git clone https://github.com/mchl18/ssl-expiration-check-bot.git
cd certchecker
make test    # Run tests
make build   # Build binary
make install # Install locally
```

## License

MIT License - see [LICENSE](LICENSE) for details. 