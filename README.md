# DNSPulse Exporter

[![Go](https://github.com/farrokhi/dnspulse_exporter/actions/workflows/go.yml/badge.svg)](https://github.com/farrokhi/dnspulse_exporter/actions/workflows/go.yml)

A Prometheus exporter for monitoring DNS query performance across multiple DNS servers and domains.

## Overview

DNSPulse Exporter performs periodic DNS queries against configured DNS servers and exposes metrics in Prometheus format. It generates randomized subdomain queries to bypass DNS caching, providing accurate real-time performance measurements.

## Features

- Monitors multiple DNS servers simultaneously
- Performs randomized subdomain queries to avoid cache hits
- Configurable query timeout and probe counts
- Exposes Prometheus metrics for query duration, success, and failure rates
- Built-in HTTP server with timeout protection against slowloris attacks
- Systemd integration for production deployment

## Build

For a specific platform:
```
env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dnspulse_exporter
```

Or for the native platform:
```
go build -ldflags "-s -w" -o dnspulse_exporter
```

## Installation

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

## Configuration

Create a configuration file at `/etc/dnspulse.yml`:

```yaml
# Prometheus listener
listen_addr: "*"
listen_port: 9953

# Log query time and success or failure - useful for debugging
verbose_logging: false

# Query timeout in milliseconds
timeout: 2500

# Domains to query (should be wildcard domains)
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

### Configuration Options

- `listen_addr`: IP address to bind the HTTP metrics server (use `*` for all interfaces)
- `listen_port`: Port for the Prometheus metrics endpoint
- `verbose_logging`: Enable detailed logging of each DNS query
- `timeout`: DNS query timeout in milliseconds
- `domains`: List of domains to query with randomized prefixes
  - `name`: Base domain name
  - `probes`: Number of queries per cycle
- `dns_servers`: List of DNS servers to monitor
  - `address`: DNS server IP address
  - `port`: DNS server port (usually 53)

## Usage

### Command Line Flags

```bash
dnspulse_exporter [flags]
```

Available flags:
- `-f <path>`: Path to configuration file (default: `/etc/dnspulse.yml`)
- `-v`: Show version information

### Running

```bash
# Using default config location
dnspulse_exporter

# Using custom config file
dnspulse_exporter -f /path/to/config.yml

# Check version
dnspulse_exporter -v
```

## Metrics

The exporter exposes the following Prometheus metrics at `/metrics`:

- `dns_query_duration_seconds`: Histogram of DNS query durations
  - Labels: `domain`, `server`
- `dns_query_success_total`: Counter of successful DNS queries
  - Labels: `domain`, `server`
- `dns_query_failures_total`: Counter of failed DNS queries
  - Labels: `domain`, `server`

### Example Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'dnspulse'
    static_configs:
      - targets: ['localhost:9953']
```

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
