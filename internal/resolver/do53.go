// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/miekg/dns"
)

// Do53Resolver implements traditional DNS over UDP or TCP (RFC 1035)
type Do53Resolver struct {
	address  string
	port     string
	useTCP   bool
	timeout  time.Duration
	client   *dns.Client
	protocol string
}

// NewDo53Resolver creates a new Do53 resolver
func NewDo53Resolver(address, port string, useTCP bool, timeout time.Duration) *Do53Resolver {
	protocol := "do53-udp"
	net := "udp"
	if useTCP {
		protocol = "do53-tcp"
		net = "tcp"
	}

	client := &dns.Client{
		Net:     net,
		Timeout: timeout,
	}

	return &Do53Resolver{
		address:  address,
		port:     port,
		useTCP:   useTCP,
		timeout:  timeout,
		client:   client,
		protocol: protocol,
	}
}

// Query performs a DNS query using Do53
func (r *Do53Resolver) Query(ctx context.Context, hostname string, qtype uint16) QueryResult {
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
func (r *Do53Resolver) Protocol() string {
	return r.protocol
}

// Close releases resources (no-op for Do53)
func (r *Do53Resolver) Close() error {
	return nil
}
