// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package resolver

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/miekg/dns"
	"github.com/quic-go/quic-go/http3"
)

// DoH3Resolver implements DNS over HTTPS using HTTP/3 (QUIC)
type DoH3Resolver struct {
	url          string
	host         string // HTTP Host header (serverName for virtual hosting)
	timeout      time.Duration
	httpClient   *http.Client
	roundTripper *http3.Transport
}

// NewDoH3Resolver creates a new DoH3 resolver
func NewDoH3Resolver(address, port, serverName string, insecureSkipVerify bool, timeout time.Duration) *DoH3Resolver {
	tlsConfig := &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: insecureSkipVerify,
	}

	roundTripper := &http3.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{
		Transport: roundTripper,
		Timeout:   timeout,
	}

	url := fmt.Sprintf("https://%s:%s/dns-query", address, port)

	return &DoH3Resolver{
		url:          url,
		host:         serverName,
		timeout:      timeout,
		httpClient:   httpClient,
		roundTripper: roundTripper,
	}
}

// Query performs a DNS query using DoH3 (RFC 8484 over HTTP/3)
func (r *DoH3Resolver) Query(ctx context.Context, hostname string, qtype uint16) QueryResult {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), qtype)

	wireMsg, err := msg.Pack()
	if err != nil {
		return QueryResult{Err: fmt.Errorf("failed to pack DNS message: %w", err)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, bytes.NewReader(wireMsg))
	if err != nil {
		return QueryResult{Err: fmt.Errorf("failed to create HTTP request: %w", err)}
	}

	req.Host = r.host // Override Host header for virtual hosting
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	start := time.Now()
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("HTTP/3 request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("HTTP status %d", resp.StatusCode),
		}
	}

	body, err := io.ReadAll(resp.Body)
	duration := time.Since(start)
	if err != nil {
		return QueryResult{
			Duration: duration,
			Err:      fmt.Errorf("failed to read response body: %w", err),
		}
	}

	response := new(dns.Msg)
	if err := response.Unpack(body); err != nil {
		return QueryResult{
			Duration: duration,
			Err:      fmt.Errorf("failed to unpack DNS response: %w", err),
		}
	}

	return QueryResult{
		Response: response,
		Duration: duration,
	}
}

// Protocol returns the protocol identifier
func (r *DoH3Resolver) Protocol() string {
	return "doh3"
}

// Close releases resources
func (r *DoH3Resolver) Close() error {
	r.httpClient.CloseIdleConnections()
	return r.roundTripper.Close()
}
