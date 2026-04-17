# Telecom Platform CLI

A comprehensive command-line interface for managing the Telecom Platform, built with Go and enhanced with Bubbletea for interactive TUI features.

## Features

- **Subscriber Management**: Create, read, update, and delete subscribers
- **Service Management**: Monitor and manage platform services
- **Billing Operations**: Handle invoices, payments, and billing reports
- **Configuration Management**: Manage CLI profiles and settings
- **Monitoring**: Real-time service health and metrics
- **Enhanced UI**: Beautiful terminal interface with colors and icons
- **Multi-format Output**: Support for table, JSON, and CSV formats
- **Plugin System**: Extensible architecture for custom commands

## Installation

### From Source
```bash
git clone https://github.com/nutcas3/telecom-platform.git
cd telecom-platform/apps/cli
go build -o telecom-cli main.go
sudo mv telecom-cli /usr/local/bin/
```

### Using Go Install
```bash
go install github.com/nutcas3/telecom-platform/apps/cli@latest
```

### From Binary Release
```bash
# Download the latest release for your platform
curl -L https://github.com/nutcas3/telecom-platform/releases/latest/download/telecom-cli-linux-amd64.tar.gz | tar xz
sudo mv telecom-cli /usr/local/bin/
```

## Quick Start

### Initial Setup
```bash
# Configure the CLI
telecom-cli config set api-endpoint http://localhost:8000
telecom-cli config set api-token your-api-token

# Test connection
telecom-cli health

# List subscribers
telecom-cli subscribers list
```

### Interactive Dashboard
```bash
# Launch the interactive dashboard
telecom-cli dashboard
```

## Commands Reference

### Global Options

| Option | Short | Description |
|--------|-------|-------------|
| `--help` | `-h` | Show help for command |
| `--version` | `-v` | Show version information |
| `--verbose` | `-V` | Enable verbose output |
| `--no-color` | | Disable colored output |
| `--format` | `-f` | Output format (table, json, csv) |
| `--profile` | `-p` | Use specific profile |

### Configuration

#### `config`
Manage CLI configuration and profiles.

```bash
# Show current configuration
telecom-cli config show

# Set configuration values
telecom-cli config set api-endpoint http://api.telecom.local
telecom-cli config set api-token your-token
telecom-cli config set timeout 30

# Get configuration values
telecom-cli config get api-endpoint

# List all profiles
telecom-cli config profiles list

# Create new profile
telecom-cli config profiles create production
telecom-cli config profiles set production.api-endpoint https://api.telecom.com

# Switch profiles
telecom-cli config profiles use production

# Delete profile
telecom-cli config profiles delete staging
```

#### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `api-endpoint` | API server URL | `http://localhost:8000` |
| `api-token` | Authentication token | `""` |
| `timeout` | Request timeout in seconds | `30` |
| `theme` | UI theme | `default` |
| `colors` | Enable colored output | `true` |
| `pager` | Use pager for long output | `true` |

### Subscribers

#### `subscribers`
Manage platform subscribers.

```bash
# List all subscribers
telecom-cli subscribers list

# List with filters
telecom-cli subscribers list --status active
telecom-cli subscribers list --limit 10
telecom-cli subscribers list --format json

# Show subscriber details
telecom-cli subscribers show 123

# Create new subscriber
telecom-cli subscribers create \
  --name "John Doe" \
  --email "john@example.com" \
  --phone "+1234567890"

# Update subscriber
telecom-cli subscribers update 123 \
  --name "John Smith" \
  --email "john.smith@example.com"

# Delete subscriber
telecom-cli subscribers delete 123

# Enhanced listing with styling
telecom-cli subscribers list --enhanced
```

#### Subscriber Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Subscriber name |
| `email` | string | Yes | Email address |
| `phone` | string | Yes | Phone number |
| `status` | string | No | Account status (active, inactive, suspended) |
| `balance` | float | No | Current balance |

### Services

#### `services`
Monitor and manage platform services.

```bash
# List all services
telecom-cli services list

# Show service details
telecom-cli services show api-server

# Enhanced service monitoring
telecom-cli services list --enhanced

# Service health check
telecom-cli services health

# Service metrics
telecom-cli services metrics --service api-server

# Restart service
telecom-cli services restart api-server
```

#### Service Status

| Status | Description |
|--------|-------------|
| `running` | Service is operational |
| `stopped` | Service is not running |
| `degraded` | Service is running with issues |
| `maintenance` | Service is under maintenance |

### Billing

#### `billing`
Manage billing operations and invoices.

```bash
# List invoices
telecom-cli billing invoices

# Show invoice details
telecom-cli billing invoices show INV-001

# Generate invoice
telecom-cli billing generate --subscriber 123 --amount 100.50

# Process payment
telecom-cli billing pay INV-001 --method credit_card

# Billing reports
telecom-cli billing reports --month 2023-12
telecom-cli billing reports --subscriber 123 --year 2023

# Enhanced billing view
telecom-cli billing invoices --enhanced
```

#### Payment Methods

| Method | Description |
|--------|-------------|
| `credit_card` | Credit card payment |
| `bank_transfer` | Bank transfer |
| `paypal` | PayPal payment |
| `crypto` | Cryptocurrency payment |

### Monitoring

#### `monitoring`
Real-time monitoring and metrics.

```bash
# System health
telecom-cli monitoring health

# Service metrics
telecom-cli monitoring metrics --service api-server

# Resource usage
telecom-cli monitoring resources

# Alert status
telecom-cli monitoring alerts

# Logs
telecom-cli monitoring logs --service api-server --tail 100

# Real-time monitoring
telecom-cli monitoring watch --refresh 5
```

#### Metrics Categories

| Category | Description |
|----------|-------------|
| `cpu` | CPU usage percentage |
| `memory` | Memory usage in MB |
| `network` | Network I/O |
| `storage` | Disk usage |
| `requests` | HTTP request metrics |
| `errors` | Error rates |

### Deployment

#### `deploy`
Deploy and manage platform deployments.

```bash
# Deploy to environment
telecom-cli deploy --environment staging

# Check deployment status
telecom-cli deploy status --environment staging

# Rollback deployment
telecom-cli deploy rollback --environment staging --version v1.0.1

# List deployments
telecom-cli deploy list

# Deployment history
telecom-cli deploy history --environment production
```

### Dashboard

#### `dashboard`
Launch interactive TUI dashboard.

```bash
# Start dashboard
telecom-cli dashboard

# Dashboard with specific theme
telecom-cli dashboard --theme dark

# Dashboard with auto-refresh
telecom-cli dashboard --refresh 10
```

#### Dashboard Features

- **Service Status**: Real-time service health
- **Metrics Charts**: CPU, memory, and network usage
- **Subscriber Overview**: Active subscribers and statistics
- **Billing Summary**: Revenue and payment status
- **Logs Viewer**: Real-time log streaming
- **Alerts Panel**: Active alerts and notifications

## Output Formats

### Table Format (Default)
```bash
telecom-cli subscribers list
```
```
+----+------------+-------------------+--------------+--------+
| ID | Name       | Email             | Phone        | Status |
+----+------------+-------------------+--------------+--------+
| 1  | John Doe   | john@example.com  | +1234567890  | active |
| 2  | Jane Smith | jane@example.com  | +0987654321  | active |
+----+------------+-------------------+--------------+--------+
```

### JSON Format
```bash
telecom-cli subscribers list --format json
```
```json
{
  "subscribers": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+1234567890",
      "status": "active",
      "balance": 100.50
    }
  ],
  "total": 1
}
```

### CSV Format
```bash
telecom-cli subscribers list --format csv
```
```csv
id,name,email,phone,status,balance
1,John Doe,john@example.com,+1234567890,active,100.50
```

## Profiles

### Profile Management

Profiles allow you to switch between different configurations (development, staging, production).

```bash
# Create production profile
telecom-cli config profiles create production
telecom-cli config profiles set production.api-endpoint https://api.telecom.com
telecom-cli config profiles set production.api-token prod-token

# Switch to production profile
telecom-cli config profiles use production

# List profiles
telecom-cli config profiles list
```

### Profile Structure

```yaml
production:
  api:
    endpoint: https://api.telecom.com
    token: prod-token
    timeout: 60
  ui:
    theme: dark
    colors: true
    pager: false
  logging:
    level: warn
    format: json
```

## Themes

### Available Themes

| Theme | Description |
|-------|-------------|
| `default` | Standard terminal colors |
| `dark` | Dark theme for better contrast |
| `light` | Light theme for bright terminals |
| `minimal` | Minimal styling without colors |

### Custom Themes

Create custom themes by modifying the configuration:

```bash
telecom-cli config set ui.theme custom
telecom-cli config set ui.colors.primary blue
telecom-cli config set ui.colors.success green
telecom-cli config set ui.colors.error red
```

## Plugins

### Installing Plugins

```bash
# Install plugin from repository
telecom-cli plugins install telecom-cli-aws

# Install from local file
telecom-cli plugins install ./custom-plugin.so

# List installed plugins
telecom-cli plugins list

# Enable/disable plugins
telecom-cli plugins enable aws
telecom-cli plugins disable aws
```

### Creating Plugins

Create custom plugins using the plugin API:

```go
package main

import (
    "github.com/nutcas3/telecom-platform/apps/cli/internal/plugin"
)

type AWSPlugin struct{}

func (p *AWSPlugin) Name() string {
    return "aws"
}

func (p *AWSPlugin) Commands() []plugin.Command {
    return []plugin.Command{
        {
            Name:        "deploy",
            Description: "Deploy to AWS",
            Handler:     p.deploy,
        },
    }
}

func (p *AWSPlugin) deploy(args []string) error {
    // Plugin implementation
    return nil
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TELECOM_API_ENDPOINT` | API server URL | `""` |
| `TELECOM_API_TOKEN` | Authentication token | `""` |
| `TELECOM_CONFIG_FILE` | Configuration file path | `~/.telecom/config.yaml` |
| `TELECOM_NO_COLOR` | Disable colors | `false` |
| `TELECOM_VERBOSE` | Enable verbose output | `false` |

## Configuration File

The CLI configuration is stored in `~/.telecom/config.yaml`:

```yaml
profiles:
  default:
    api:
      endpoint: http://localhost:8000
      token: ""
      timeout: 30
    ui:
      theme: default
      colors: true
      pager: true
    logging:
      level: info
      format: text

current_profile: default
plugins:
  enabled: []
  disabled: []
```

## Examples

### Common Workflows

#### 1. Daily Operations
```bash
# Check system health
telecom-cli health

# Review new subscribers
telecom-cli subscribers list --status new --enhanced

# Check service status
telecom-cli services list --enhanced

# Review billing
telecom-cli billing invoices --unpaid
```

#### 2. Troubleshooting
```bash
# Check service logs
telecom-cli monitoring logs --service api-server --tail 100

# View metrics
telecom-cli monitoring metrics --service api-server

# Check system resources
telecom-cli monitoring resources
```

#### 3. Deployment
```bash
# Deploy to staging
telecom-cli deploy --environment staging

# Monitor deployment
telecom-cli deploy status --environment staging

# Check health after deployment
telecom-cli health --environment staging
```

### Advanced Usage

#### Batch Operations
```bash
# Export subscribers to CSV
telecom-cli subscribers list --format csv > subscribers.csv

# Bulk update subscribers
while read line; do
  telecom-cli subscribers update $line
done < updates.txt
```

#### Automation Scripts
```bash
#!/bin/bash
# health-check.sh

# Check system health
if ! telecom-cli health; then
  echo "System unhealthy!" >&2
  exit 1
fi

# Check critical services
services=("api-server" "charging-engine" "packet-gateway")
for service in "${services[@]}"; do
  if ! telecom-cli services show $service | grep -q "running"; then
    echo "Service $service not running!" >&2
    exit 1
  fi
done

echo "All systems healthy"
```

## Troubleshooting

### Common Issues

#### Connection Issues
```bash
# Check API endpoint
telecom-cli config get api-endpoint

# Test connection
telecom-cli health --verbose

# Check timeout settings
telecom-cli config get timeout
```

#### Authentication Issues
```bash
# Verify API token
telecom-cli config get api-token

# Test with different token
TELECOM_API_TOKEN=new-token telecom-cli health
```

#### Performance Issues
```bash
# Check response times
telecom-cli health --verbose

# Use JSON format for faster processing
telecom-cli subscribers list --format json
```

### Debug Mode

Enable debug mode for detailed troubleshooting:

```bash
# Enable verbose output
telecom-cli --verbose subscribers list

# Set debug log level
telecom-cli config set logging.level debug

# Enable request tracing
telecom-cli config set api.trace true
```

## Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/nutcas3/telecom-platform.git
cd telecom-platform/apps/cli

# Install dependencies
go mod download

# Run tests
go test ./...

# Build development version
go build -o telecom-cli main.go
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run integration tests
go test ./tests/... -v

# Run benchmarks
go test ./tests/... -bench=.

# Test coverage
go test ./... -cover
```

### Code Style

The project uses the standard Go formatting tools:

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run static analysis
go vet ./...
```

## Support

- **Documentation**: [https://docs.telecom-platform.com](https://docs.telecom-platform.com)
- **Issues**: [GitHub Issues](https://github.com/nutcas3/telecom-platform/issues)
- **Discussions**: [GitHub Discussions](https://github.com/nutcas3/telecom-platform/discussions)
- **Community**: [Discord Server](https://discord.gg/telecom-platform)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v1.0.0 (2023-12-01)
- Initial release
- Core subscriber and service management
- Enhanced UI with Bubbletea
- Multi-format output support
- Profile management
- Plugin system
- Interactive dashboard

### v1.1.0 (2023-12-15)
- Added billing commands
- Enhanced monitoring features
- Improved error handling
- Performance optimizations

### v1.2.0 (2024-01-01)
- Added deployment commands
- Enhanced plugin system
- New themes and styling options
- Better configuration management
