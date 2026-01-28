// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2026 Babak Farrokhi

package resolver

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/miekg/dns"
)

// DoTResolver implements DNS over TLS (RFC 7858)
type DoTResolver struct {
	address   string
	port      string
	timeout   time.Duration
	client    *dns.Client
	tlsConfig *tls.Config
}

// NewDoTResolver creates a new DoT resolver
func NewDoTResolver(address, port, serverName string, insecureSkipVerify bool, timeout time.Duration) *DoTResolver {
	tlsConfig := &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: insecureSkipVerify,
	}

	client := &dns.Client{
		Net:       "tcp-tls",
		Timeout:   timeout,
		TLSConfig: tlsConfig,
	}

	return &DoTResolver{
		address:   address,
		port:      port,
		timeout:   timeout,
		client:    client,
		tlsConfig: tlsConfig,
	}
}

// Query performs a DNS query using DoT
func (r *DoTResolver) Query(ctx context.Context, hostname string, qtype uint16) QueryResult {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), qtype)

	serverAddr := fmt.Sprintf("%s:%s", r.address, r.port)

	start := time.Now()
	resp, _, err := r.client.ExchangeContext(ctx, msg, serverAddr)
	duration := time.Since(start)

	return QueryResult{
		Response: resp,
		Duration: duration,
		Err:      err,
	}
}

// Protocol returns the protocol identifier
func (r *DoTResolver) Protocol() string {
	return "dot"
}

// Close releases resources (no-op for DoT)
func (r *DoTResolver) Close() error {
	return nil
}
