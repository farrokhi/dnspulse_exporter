// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2026 Babak Farrokhi

package resolver

import (
	"context"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestDo53ResolverProtocol(t *testing.T) {
	t.Run("UDP protocol", func(t *testing.T) {
		r := NewDo53Resolver("8.8.8.8", "53", false, 2*time.Second)
		if r.Protocol() != "do53-udp" {
			t.Errorf("Expected 'do53-udp', got '%s'", r.Protocol())
		}
	})

	t.Run("TCP protocol", func(t *testing.T) {
		r := NewDo53Resolver("8.8.8.8", "53", true, 2*time.Second)
		if r.Protocol() != "do53-tcp" {
			t.Errorf("Expected 'do53-tcp', got '%s'", r.Protocol())
		}
	})
}

func TestDoTResolverProtocol(t *testing.T) {
	r := NewDoTResolver("1.1.1.1", "853", "cloudflare-dns.com", false, 2*time.Second)
	if r.Protocol() != "dot" {
		t.Errorf("Expected 'dot', got '%s'", r.Protocol())
	}
}

func TestDoHResolverProtocol(t *testing.T) {
	r := NewDoHResolver("dns.google", "443", "dns.google", false, 2*time.Second)
	if r.Protocol() != "doh" {
		t.Errorf("Expected 'doh', got '%s'", r.Protocol())
	}
}

func TestDoH3ResolverProtocol(t *testing.T) {
	r := NewDoH3Resolver("dns.google", "443", "dns.google", false, 2*time.Second)
	if r.Protocol() != "doh3" {
		t.Errorf("Expected 'doh3', got '%s'", r.Protocol())
	}
}

func TestDoQResolverProtocol(t *testing.T) {
	r := NewDoQResolver("dns.adguard-dns.com", "853", "dns.adguard-dns.com", false, 2*time.Second)
	if r.Protocol() != "doq" {
		t.Errorf("Expected 'doq', got '%s'", r.Protocol())
	}
}

func TestDo53Query(t *testing.T) {
	r := NewDo53Resolver("8.8.8.8", "53", false, 5*time.Second)
	defer r.Close()

	ctx := context.Background()
	result := r.Query(ctx, "example.com", dns.TypeA)

	if result.Duration < 0 {
		t.Errorf("Expected non-negative duration, got %v", result.Duration)
	}

	if result.Err != nil {
		t.Logf("DNS query failed (may be expected in some environments): %v", result.Err)
		return
	}

	if result.Response == nil {
		t.Error("Expected response, got nil")
	}
}

func TestDo53QueryTimeout(t *testing.T) {
	r := NewDo53Resolver("192.0.2.1", "53", false, 100*time.Millisecond)
	defer r.Close()

	ctx := context.Background()
	result := r.Query(ctx, "example.com", dns.TypeA)

	if result.Err == nil {
		t.Error("Expected error for non-routable address")
	}

	if result.Duration < 0 {
		t.Errorf("Expected non-negative duration, got %v", result.Duration)
	}
}

func TestResolverClose(t *testing.T) {
	resolvers := []Resolver{
		NewDo53Resolver("8.8.8.8", "53", false, 2*time.Second),
		NewDo53Resolver("8.8.8.8", "53", true, 2*time.Second),
		NewDoTResolver("1.1.1.1", "853", "cloudflare-dns.com", false, 2*time.Second),
		NewDoHResolver("dns.google", "443", "dns.google", false, 2*time.Second),
		NewDoH3Resolver("dns.google", "443", "dns.google", false, 2*time.Second),
		NewDoQResolver("dns.adguard-dns.com", "853", "dns.adguard-dns.com", false, 2*time.Second),
	}

	for _, r := range resolvers {
		if err := r.Close(); err != nil {
			t.Errorf("Close() returned error for %s: %v", r.Protocol(), err)
		}
	}
}
