// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2026 Babak Farrokhi

package resolver

import (
	"testing"
	"time"

	"dnspulse_exporter/internal/config"
)

func TestNewResolver(t *testing.T) {
	timeout := 2 * time.Second

	tests := []struct {
		name          string
		server        config.DNSServer
		expectedProto string
		expectError   bool
	}{
		{
			name: "do53-udp",
			server: config.DNSServer{
				Address:  "8.8.8.8",
				Port:     "53",
				Protocol: config.ProtocolDo53UDP,
			},
			expectedProto: "do53-udp",
		},
		{
			name: "do53-tcp",
			server: config.DNSServer{
				Address:  "8.8.8.8",
				Port:     "53",
				Protocol: config.ProtocolDo53TCP,
			},
			expectedProto: "do53-tcp",
		},
		{
			name: "dot",
			server: config.DNSServer{
				Address:  "1.1.1.1",
				Port:     "853",
				Protocol: config.ProtocolDoT,
				TLS:      &config.TLSConfig{ServerName: "cloudflare-dns.com"},
			},
			expectedProto: "dot",
		},
		{
			name: "doh",
			server: config.DNSServer{
				Address:  "dns.google",
				Port:     "443",
				Protocol: config.ProtocolDoH,
				TLS:      &config.TLSConfig{ServerName: "dns.google"},
			},
			expectedProto: "doh",
		},
		{
			name: "doh3",
			server: config.DNSServer{
				Address:  "dns.google",
				Port:     "443",
				Protocol: config.ProtocolDoH3,
				TLS:      &config.TLSConfig{ServerName: "dns.google"},
			},
			expectedProto: "doh3",
		},
		{
			name: "doq",
			server: config.DNSServer{
				Address:  "dns.adguard-dns.com",
				Port:     "853",
				Protocol: config.ProtocolDoQ,
				TLS:      &config.TLSConfig{ServerName: "dns.adguard-dns.com"},
			},
			expectedProto: "doq",
		},
		{
			name: "unsupported protocol",
			server: config.DNSServer{
				Address:  "8.8.8.8",
				Port:     "53",
				Protocol: "unknown",
			},
			expectError: true,
		},
		{
			name: "empty protocol rejected",
			server: config.DNSServer{
				Address:  "8.8.8.8",
				Port:     "53",
				Protocol: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewResolver(tt.server, timeout)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewResolver failed: %v", err)
			}

			if r.Protocol() != tt.expectedProto {
				t.Errorf("Expected protocol '%s', got '%s'", tt.expectedProto, r.Protocol())
			}

			_ = r.Close()
		})
	}
}

func TestNewResolverTLSDefaults(t *testing.T) {
	timeout := 2 * time.Second

	server := config.DNSServer{
		Address:  "dns.google",
		Port:     "853",
		Protocol: config.ProtocolDoT,
	}

	r, err := NewResolver(server, timeout)
	if err != nil {
		t.Fatalf("NewResolver failed: %v", err)
	}
	defer func() { _ = r.Close() }()

	if r.Protocol() != "dot" {
		t.Errorf("Expected protocol 'dot', got '%s'", r.Protocol())
	}
}
