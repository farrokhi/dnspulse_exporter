// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package resolver

import (
	"context"
	"time"

	"github.com/miekg/dns"
)

// QueryResult contains the result of a DNS query
type QueryResult struct {
	Response *dns.Msg
	Duration time.Duration
	Err      error
}

// Resolver is the interface that all DNS resolvers must implement
type Resolver interface {
	// Query performs a DNS query for the given hostname and record type
	Query(ctx context.Context, hostname string, qtype uint16) QueryResult

	// Protocol returns the protocol identifier (e.g., "do53-udp", "dot", "doh")
	Protocol() string

	// Close releases any resources held by the resolver
	Close() error
}
