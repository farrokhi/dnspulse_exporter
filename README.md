# DNS Pulse Exporter

[![Go](https://github.com/farrokhi/dnspulse_exporter/actions/workflows/go.yml/badge.svg)](https://github.com/farrokhi/dnspulse_exporter/actions/workflows/go.yml)

A Prometheus exporter for monitoring DNS server performance and availability. This tool continuously performs DNS queries against multiple servers and domains, measuring response times and success rates to help you monitor DNS infrastructure health.

## Features

- **DNS Performance Monitoring**: Measures query response times and success/failure rates
- **Cache Bypass**: Uses randomized hostname prefixes to avoid DNS caching effects
- **Multi-Server Support**: Tests multiple DNS servers sequentially
- **Prometheus Integration**: Exposes metrics in Prometheus format via `/metrics` endpoint
- **Configurable Testing**: Customizable domains, probe counts, timeouts, and server lists
- **Continuous Monitoring**: Runs automated tests every 30 seconds
- **Verbose Logging**: Optional detailed logging for debugging
- **Security Hardening**: HTTP server with timeout protection against slowloris attacks
- **Systemd Integration**: Ready for production deployment

## Metrics Exported

- `dns_query_duration_seconds` - Histogram of DNS query response times
- `dns_query_success_total` - Counter of successful DNS queries
- `dns_query_failures_total` - Counter of failed DNS queries

All metrics include labels for `domain` and `server` to enable detailed analysis.

## Installation

### Using Make (recommended)

```bash
make build       # Build the binary
make test        # Run tests
make install     # Install to system
make help        # Show all available targets
```

The `make install` command will install:
- Binary to `/usr/local/bin/dnspulse_exporter`
- Configuration file to `/etc/dnspulse.yml`
- Systemd service to `/etc/systemd/system/dnspulse.service`

After installation with systemd:
```bash
sudo systemctl daemon-reload
sudo systemctl enable dnspulse.service
sudo systemctl start dnspulse.service
```

### Building from Source

For a specific platform:
```bash
env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dnspulse_exporter
```

Or for the native platform:
```bash
go build -ldflags "-s -w" -o dnspulse_exporter
```

### Manual Installation

```bash
# Copy binary
sudo cp dnspulse_exporter /usr/bin/

# Copy configuration
sudo cp dnspulse.yml /etc/

# Install systemd service (optional)
sudo cp systemd/dnspulse.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable dnspulse.service
sudo systemctl start dnspulse.service
```

## Running

```bash
# Use default config file (/etc/dnspulse.yml)
./dnspulse_exporter

# Specify custom config file
./dnspulse_exporter -f /path/to/config.yml

# Show version
./dnspulse_exporter -v
```

The exporter will start an HTTP server on the configured port (default: 9953) and begin monitoring DNS servers.

## Configuration

Create a YAML configuration file (default: `/etc/dnspulse.yml`) with the following structure:

### Basic Configuration Example

```yaml
# Prometheus metrics listener
listen_addr: "*"
listen_port: 9953

# Enable detailed logging (useful for debugging)
verbose_logging: false

# Query timeout in milliseconds (default: 2000 if not specified)
timeout: 2500

# Domains to test (use wildcard domains for cache bypass)
domains:
  - name: "blogspot.com"
    probes: 3
  - name: "wordpress.com"
    probes: 3

# DNS servers to monitor
dns_servers:
  - address: "8.8.8.8"
    port: "53"
  - address: "9.9.9.9"
    port: "53"
  - address: "1.1.1.1"
    port: "53"
```

### Advanced Configuration Example

```yaml
listen_addr: "0.0.0.0"
listen_port: 9953
verbose_logging: true
timeout: 5000

domains:
  - name: "example.com"
    probes: 5
  - name: "github.com"
    probes: 3
  - name: "stackoverflow.com"
    probes: 2

dns_servers:
  # Public DNS servers
  - address: "8.8.8.8"
    port: "53"
  - address: "8.8.4.4"
    port: "53"
  - address: "1.1.1.1"
    port: "53"
  - address: "1.0.0.1"
    port: "53"
  # Internal corporate DNS
  - address: "192.168.1.10"
    port: "53"
  - address: "10.0.0.53"
    port: "53"
```

**Note**: This tool supports standard DNS queries only (UDP/TCP on port 53). DNS over HTTPS (DoH) and DNS over TLS (DoT) are not supported.

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `listen_addr` | IP address to bind the metrics server | `"*"` (all interfaces) |
| `listen_port` | Port for the metrics server | `9953` |
| `verbose_logging` | Enable detailed query logging | `false` |
| `timeout` | DNS query timeout in milliseconds | `2000` (default if not specified) |
| `domains` | List of domains to test with probe counts | - |
| `dns_servers` | List of DNS servers with address and port | - |

## Systemd Service

A systemd service file is included in the `systemd/` directory. To install:

```bash
sudo cp systemd/dnspulse.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable dnspulse
sudo systemctl start dnspulse
```

## Prometheus Configuration

Add the following to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'dns-pulse'
    static_configs:
      - targets: ['localhost:9953']
    scrape_interval: 30s
```

## How It Works

1. **Randomized Testing**: For each domain, the exporter generates a random prefix (e.g., `abc12.example.com`) to bypass DNS caching
2. **Sequential Queries**: Tests all configured DNS servers sequentially, with a 500ms delay between each probe
3. **Metrics Collection**: Records response times, success/failure counts with server and domain labels
4. **Continuous Operation**: Repeats the testing cycle every 30 seconds
5. **Prometheus Export**: Makes metrics available at `/metrics` endpoint for scraping

## Security

The HTTP server includes the following timeout protections:
- **ReadTimeout**: 5 seconds - protects against slowloris attacks
- **WriteTimeout**: 10 seconds - ensures timely response delivery
- **IdleTimeout**: 120 seconds - manages keep-alive connections

When running via systemd, the service includes security hardening options such as:
- No new privileges
- Protected system and home directories
- Private temporary directory
- Restricted address families (IPv4/IPv6 only)

## License

This project is licensed under the BSD 2-Clause License. See the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Babak Farrokhi
