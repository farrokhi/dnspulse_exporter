// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package prober

import (
	"testing"

	"dnspulse_exporter/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Domains: []config.Domain{
			{Name: "example.com", Probes: 1},
		},
		DNSServers: []config.DNSServer{
			{Address: "8.8.8.8", Port: "53", Protocol: config.ProtocolDo53UDP},
		},
		Timeout: 2000,
	}

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer p.Close()

	if len(p.resolvers) != 1 {
		t.Errorf("Expected 1 resolver, got %d", len(p.resolvers))
	}
}

func TestNewWithMultipleServers(t *testing.T) {
	cfg := &config.Config{
		Domains: []config.Domain{
			{Name: "example.com", Probes: 1},
		},
		DNSServers: []config.DNSServer{
			{Address: "8.8.8.8", Port: "53", Protocol: config.ProtocolDo53UDP},
			{Address: "1.1.1.1", Port: "53", Protocol: config.ProtocolDo53TCP},
		},
		Timeout: 2000,
	}

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer p.Close()

	if len(p.resolvers) != 2 {
		t.Errorf("Expected 2 resolvers, got %d", len(p.resolvers))
	}
}

func TestNewWithInvalidProtocol(t *testing.T) {
	cfg := &config.Config{
		Domains: []config.Domain{
			{Name: "example.com", Probes: 1},
		},
		DNSServers: []config.DNSServer{
			{Address: "8.8.8.8", Port: "53", Protocol: "invalid"},
		},
		Timeout: 2000,
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("Expected error for invalid protocol, got nil")
	}
}

func TestGenerateRandomPrefix(t *testing.T) {
	t.Run("generates prefix of correct length", func(t *testing.T) {
		prefix := generateRandomPrefix(5)
		if len(prefix) != 8 {
			t.Errorf("Expected prefix length 8, got %d", len(prefix))
		}
	})

	t.Run("generates unique prefixes", func(t *testing.T) {
		prefix1 := generateRandomPrefix(5)
		prefix2 := generateRandomPrefix(5)
		if prefix1 == prefix2 {
			t.Error("Expected different prefixes, got identical ones")
		}
	})

	t.Run("generates valid base32 string", func(t *testing.T) {
		prefix := generateRandomPrefix(5)
		for _, c := range prefix {
			if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7')) {
				t.Errorf("Invalid base32 character: %c", c)
			}
		}
	})
}

func TestDefaultTimeout(t *testing.T) {
	cfg := &config.Config{
		Domains: []config.Domain{
			{Name: "example.com", Probes: 1},
		},
		DNSServers: []config.DNSServer{
			{Address: "8.8.8.8", Port: "53", Protocol: config.ProtocolDo53UDP},
		},
		Timeout: 0,
	}

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer p.Close()
}
