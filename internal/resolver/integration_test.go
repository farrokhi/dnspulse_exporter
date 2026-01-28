// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

//go:build integration

package resolver

import (
	"context"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// Integration tests against Quad9 (9.9.9.9) which supports all protocols.
// Run with: go test -tags=integration -v ./internal/resolver/

const (
	quad9IP         = "9.9.9.9"
	quad9ServerName = "dns.quad9.net"
	testDomain      = "example.com"
	testTimeout     = 10 * time.Second
)

func TestIntegrationDo53UDP(t *testing.T) {
	r := NewDo53Resolver(quad9IP, "53", false, testTimeout)
	defer r.Close()

	ctx := context.Background()
	result := r.Query(ctx, testDomain, dns.TypeA)

	if result.Err != nil {
		t.Fatalf("Do53 UDP query failed: %v", result.Err)
	}

	if result.Response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(result.Response.Answer) == 0 {
		t.Error("Expected at least one answer record")
	}

	t.Logf("Do53 UDP: %s resolved in %v", testDomain, result.Duration)
}

func TestIntegrationDo53TCP(t *testing.T) {
	r := NewDo53Resolver(quad9IP, "53", true, testTimeout)
	defer r.Close()

	ctx := context.Background()
	result := r.Query(ctx, testDomain, dns.TypeA)

	if result.Err != nil {
		t.Fatalf("Do53 TCP query failed: %v", result.Err)
	}

	if result.Response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(result.Response.Answer) == 0 {
		t.Error("Expected at least one answer record")
	}

	t.Logf("Do53 TCP: %s resolved in %v", testDomain, result.Duration)
}

func TestIntegrationDoT(t *testing.T) {
	r := NewDoTResolver(quad9IP, "853", quad9ServerName, false, testTimeout)
	defer r.Close()

	ctx := context.Background()
	result := r.Query(ctx, testDomain, dns.TypeA)

	if result.Err != nil {
		t.Fatalf("DoT query failed: %v", result.Err)
	}

	if result.Response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(result.Response.Answer) == 0 {
		t.Error("Expected at least one answer record")
	}

	t.Logf("DoT: %s resolved in %v", testDomain, result.Duration)
}

func TestIntegrationDoH(t *testing.T) {
	r := NewDoHResolver(quad9ServerName, "443", quad9ServerName, false, testTimeout)
	defer r.Close()

	ctx := context.Background()
	result := r.Query(ctx, testDomain, dns.TypeA)

	if result.Err != nil {
		t.Fatalf("DoH query failed: %v", result.Err)
	}

	if result.Response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(result.Response.Answer) == 0 {
		t.Error("Expected at least one answer record")
	}

	t.Logf("DoH: %s resolved in %v", testDomain, result.Duration)
}

func TestIntegrationDoH3(t *testing.T) {
	r := NewDoH3Resolver(quad9ServerName, "443", quad9ServerName, false, testTimeout)
	defer r.Close()

	ctx := context.Background()
	result := r.Query(ctx, testDomain, dns.TypeA)

	if result.Err != nil {
		t.Fatalf("DoH3 query failed: %v", result.Err)
	}

	if result.Response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(result.Response.Answer) == 0 {
		t.Error("Expected at least one answer record")
	}

	t.Logf("DoH3: %s resolved in %v", testDomain, result.Duration)
}

func TestIntegrationDoQ(t *testing.T) {
	r := NewDoQResolver(quad9ServerName, "853", quad9ServerName, false, testTimeout)
	defer r.Close()

	ctx := context.Background()
	result := r.Query(ctx, testDomain, dns.TypeA)

	if result.Err != nil {
		t.Fatalf("DoQ query failed: %v", result.Err)
	}

	if result.Response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(result.Response.Answer) == 0 {
		t.Error("Expected at least one answer record")
	}

	t.Logf("DoQ: %s resolved in %v", testDomain, result.Duration)
}

func TestIntegrationAllProtocols(t *testing.T) {
	tests := []struct {
		name     string
		resolver Resolver
	}{
		{"Do53-UDP", NewDo53Resolver(quad9IP, "53", false, testTimeout)},
		{"Do53-TCP", NewDo53Resolver(quad9IP, "53", true, testTimeout)},
		{"DoT", NewDoTResolver(quad9IP, "853", quad9ServerName, false, testTimeout)},
		{"DoH", NewDoHResolver(quad9ServerName, "443", quad9ServerName, false, testTimeout)},
		{"DoH3", NewDoH3Resolver(quad9ServerName, "443", quad9ServerName, false, testTimeout)},
		{"DoQ", NewDoQResolver(quad9ServerName, "853", quad9ServerName, false, testTimeout)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.resolver.Close()

			ctx := context.Background()
			result := tt.resolver.Query(ctx, testDomain, dns.TypeA)

			if result.Err != nil {
				t.Errorf("%s query failed: %v", tt.name, result.Err)
				return
			}

			if result.Response == nil {
				t.Errorf("%s: expected response, got nil", tt.name)
				return
			}

			if len(result.Response.Answer) == 0 {
				t.Errorf("%s: expected at least one answer record", tt.name)
				return
			}

			t.Logf("%s (%s): %s resolved in %v",
				tt.name, tt.resolver.Protocol(), testDomain, result.Duration)
		})
	}
}
