// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2026 Babak Farrokhi

package resolver

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/net/http2"
)

// DoHResolver implements DNS over HTTPS (RFC 8484)
type DoHResolver struct {
	url        string
	host       string // HTTP Host header (serverName for virtual hosting)
	timeout    time.Duration
	httpClient *http.Client
	transport  *http2.Transport
}

// NewDoHResolver creates a new DoH resolver using strict HTTP/2
func NewDoHResolver(address, port, serverName string, insecureSkipVerify bool, timeout time.Duration) *DoHResolver {
	tlsConfig := &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: insecureSkipVerify,
		NextProtos:         []string{"h2"},
	}

	transport := &http2.Transport{
		TLSClientConfig:    tlsConfig,
		DisableCompression: false,
		AllowHTTP:          false,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			netDialer := &net.Dialer{Timeout: timeout}
			conn, err := netDialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			tlsConn := tls.Client(conn, tlsConfig)
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				_ = conn.Close()
				return nil, err
			}
			return tlsConn, nil
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	url := fmt.Sprintf("https://%s:%s/dns-query", address, port)

	return &DoHResolver{
		url:        url,
		host:       serverName,
		timeout:    timeout,
		httpClient: httpClient,
		transport:  transport,
	}
}

// Query performs a DNS query using DoH (RFC 8484 wire format over HTTP/2)
func (r *DoHResolver) Query(ctx context.Context, hostname string, qtype uint16) QueryResult {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), qtype)
	msg.Id = 0

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
			Err:      fmt.Errorf("HTTP/2 request failed: %w", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("HTTP status %d: %s", resp.StatusCode, string(body)),
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
func (r *DoHResolver) Protocol() string {
	return "doh"
}

// Close releases resources
func (r *DoHResolver) Close() error {
	r.httpClient.CloseIdleConnections()
	r.transport.CloseIdleConnections()
	return nil
}
