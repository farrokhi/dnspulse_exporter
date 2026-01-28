# DNS Pulse Exporter

[![Go](https://github.com/farrokhi/dnspulse_exporter/actions/workflows/go.yml/badge.svg)](https://github.com/farrokhi/dnspulse_exporter/actions/workflows/go.yml)

A Prometheus exporter for monitoring DNS query performance across multiple DNS servers, domains, and protocols.

## Overview

DNSPulse Exporter performs periodic DNS queries against configured DNS servers and exposes metrics in Prometheus format. It generates randomized subdomain queries to bypass DNS caching, providing accurate real-time performance measurements.

## Features

DNSPulse supports multiple DNS protocols for comprehensive monitoring:

| Protocol | Description | Default Port | RFC |
|----------|-------------|--------------|-----|
| do53-udp | Traditional DNS over UDP | 53 | RFC 1035 |
| do53-tcp | Traditional DNS over TCP | 53 | RFC 1035 |
| dot | DNS over TLS | 853 | RFC 7858 |
| doh | DNS over HTTPS (HTTP/2) | 443 | RFC 8484 |
| doh3 | DNS over HTTPS (HTTP/3) | 443 | RFC 8484 |
| doq | DNS over QUIC | 853 | RFC 9250 |

Additional features include randomized subdomain queries to avoid cache hits, configurable timeouts and probe counts, per-protocol metrics with Prometheus labels, and systemd integration for production deployment.

## Metrics Exported

- `dns_query_duration_seconds` - Histogram of DNS query response times
- `dns_query_success_total` - Counter of successful DNS queries
- `dns_query_failures_total` - Counter of failed DNS queries

All metrics include labels for `domain`, `server`, and `protocol` to enable detailed analysis.

## Installation

Using Make (recommended):

```bash
make build       # Build the binary
make test        # Run unit tests
make test-integration  # Run integration tests against Quad9
make help        # Show all available targets
```

Manual build:

```bash
go build -ldflags "-s -w" -o dnspulse_exporter ./cmd/dnspulse_exporter
```

## System Installation

Using Make:

```bash
sudo make install
```

This installs the binary to `/usr/local/bin/dnspulse_exporter`, the configuration file to `/etc/dnspulse.yml`, and the systemd service to `/etc/systemd/system/dnspulse.service`.

After installation:

```bash
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

# Show version (displays version, git commit hash, and build time)
./dnspulse_exporter -v
```

The exporter will start an HTTP server on the configured port (default: 9953) and begin monitoring DNS servers.

## Configuration

Create a YAML configuration file (default: `/etc/dnspulse.yml`) with the following structure:

### Basic Configuration Example

```yaml
listen_addr: "*"
listen_port: 9953
verbose_logging: false
timeout: 2500

domains:
  - name: "example.com"
    probes: 3

dns_servers:
  # Traditional DNS (UDP)
  - address: "9.9.9.9"
    port: "53"
    protocol: "do53-udp"

  # Traditional DNS (TCP)
  - address: "9.9.9.9"
    port: "53"
    protocol: "do53-tcp"

  # DNS over TLS
  - address: "9.9.9.9"
    port: "853"
    protocol: "dot"
    tls:
      server_name: "dns.quad9.net"

  # DNS over HTTPS (HTTP/2)
  - address: "dns.quad9.net"
    port: "443"
    protocol: "doh"
    tls:
      server_name: "dns.quad9.net"

  # DNS over HTTPS (HTTP/3)
  - address: "dns.quad9.net"
    port: "443"
    protocol: "doh3"
    tls:
      server_name: "dns.quad9.net"

  # DNS over QUIC
  - address: "dns.quad9.net"
    port: "853"
    protocol: "doq"
    tls:
      server_name: "dns.quad9.net"
```

### Configuration Reference

Global settings:

| Field | Description | Default |
|-------|-------------|---------|
| listen_addr | IP address to bind (use `*` for all interfaces) | - |
| listen_port | Port for Prometheus metrics endpoint | - |
| verbose_logging | Enable detailed query logging | false |
| timeout | DNS query timeout in milliseconds | - |

Domain settings:

| Field | Description |
|-------|-------------|
| name | Base domain name for queries |
| probes | Number of queries per cycle |

DNS server settings:

| Field | Description | Required |
|-------|-------------|----------|
| address | DNS server IP or hostname | Yes |
| port | DNS server port | No (protocol default) |
| protocol | Protocol to use (see table above) | No (do53-udp) |
| tls.server_name | TLS SNI server name | No (uses address) |
| tls.insecure_skip_verify | Skip TLS certificate verification | No (false) |

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
```

## Prometheus Configuration

The exporter exposes the following Prometheus metrics at `/metrics`:

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| dns_query_duration_seconds | Histogram | domain, server, protocol | DNS query duration |
| dns_query_success_total | Counter | domain, server, protocol | Successful queries |
| dns_query_failures_total | Counter | domain, server, protocol | Failed queries |

Example Prometheus queries:

```promql
# Average query duration by protocol
avg by (protocol) (rate(dns_query_duration_seconds_sum[5m]) / rate(dns_query_duration_seconds_count[5m]))

# Success rate by server
sum by (server) (rate(dns_query_success_total[5m])) /
(sum by (server) (rate(dns_query_success_total[5m])) + sum by (server) (rate(dns_query_failures_total[5m])))
```

Prometheus scrape configuration:

```yaml
scrape_configs:
  - job_name: 'dns-pulse'
    static_configs:
      - targets: ['localhost:9953']
    scrape_interval: 30s
```

## Project Structure

```
dnspulse_exporter/
├── cmd/dnspulse_exporter/    # Application entry point
├── internal/
│   ├── config/               # Configuration parsing
│   ├── metrics/              # Prometheus metrics
│   ├── prober/               # Query orchestration
│   └── resolver/             # Protocol implementations
├── dnspulse.yml              # Example configuration
└── Makefile
```

## License

BSD 2-Clause License. See [LICENSE](LICENSE) for details.

Copyright (c) 2026 Babak Farrokhi
