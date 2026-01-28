// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2026 Babak Farrokhi

package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("valid config file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-config-*.yml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		configContent := `
listen_addr: "127.0.0.1"
listen_port: "9953"
verbose_logging: true
timeout: 3000
domains:
  - name: "example.com"
    probes: 2
dns_servers:
  - address: "8.8.8.8"
    port: "53"
`
		if _, err := tempFile.WriteString(configContent); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		config, err := Load(tempFile.Name())
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if config.ListenAddress != "127.0.0.1" {
			t.Errorf("Expected ListenAddress '127.0.0.1', got '%s'", config.ListenAddress)
		}
		if config.ListenPort != "9953" {
			t.Errorf("Expected ListenPort '9953', got '%s'", config.ListenPort)
		}
		if !config.VerboseLogging {
			t.Error("Expected VerboseLogging to be true")
		}
		if config.Timeout != 3000 {
			t.Errorf("Expected Timeout 3000, got %d", config.Timeout)
		}
		if len(config.Domains) != 1 {
			t.Errorf("Expected 1 domain, got %d", len(config.Domains))
		}
		if len(config.DNSServers) != 1 {
			t.Errorf("Expected 1 DNS server, got %d", len(config.DNSServers))
		}
	})

	t.Run("non-existent config file", func(t *testing.T) {
		_, err := Load("/nonexistent/config.yml")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-invalid-*.yml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		invalidContent := "invalid: yaml: content: ["
		if _, err := tempFile.WriteString(invalidContent); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		_, err = Load(tempFile.Name())
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})
}

func TestDefaultProtocol(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-config-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	configContent := `
listen_addr: "127.0.0.1"
listen_port: "9953"
domains:
  - name: "example.com"
    probes: 1
dns_servers:
  - address: "8.8.8.8"
    port: "53"
`
	if _, err := tempFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	config, err := Load(tempFile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if config.DNSServers[0].Protocol != ProtocolDo53UDP {
		t.Errorf("Expected default protocol '%s', got '%s'", ProtocolDo53UDP, config.DNSServers[0].Protocol)
	}
}

func TestProtocolValidation(t *testing.T) {
	t.Run("valid protocols", func(t *testing.T) {
		for proto := range ValidProtocols {
			tempFile, err := os.CreateTemp("", "test-config-*.yml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			configContent := `
listen_addr: "127.0.0.1"
listen_port: "9953"
domains:
  - name: "example.com"
    probes: 1
dns_servers:
  - address: "8.8.8.8"
    port: "53"
    protocol: "` + proto + `"
`
			if _, err := tempFile.WriteString(configContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tempFile.Close()

			_, err = Load(tempFile.Name())
			if err != nil {
				t.Errorf("Expected no error for valid protocol '%s', got: %v", proto, err)
			}
		}
	})

	t.Run("invalid protocol", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-config-*.yml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		configContent := `
listen_addr: "127.0.0.1"
listen_port: "9953"
domains:
  - name: "example.com"
    probes: 1
dns_servers:
  - address: "8.8.8.8"
    port: "53"
    protocol: "invalid-protocol"
`
		if _, err := tempFile.WriteString(configContent); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		_, err = Load(tempFile.Name())
		if err == nil {
			t.Error("Expected error for invalid protocol, got nil")
		}
	})
}

func TestTLSConfigDefaults(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-config-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	configContent := `
listen_addr: "127.0.0.1"
listen_port: "9953"
domains:
  - name: "example.com"
    probes: 1
dns_servers:
  - address: "dns.google"
    port: "853"
    protocol: "dot"
`
	if _, err := tempFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	config, err := Load(tempFile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if config.DNSServers[0].TLS == nil {
		t.Fatal("Expected TLS config to be created")
	}
	if config.DNSServers[0].TLS.ServerName != "dns.google" {
		t.Errorf("Expected server_name 'dns.google', got '%s'", config.DNSServers[0].TLS.ServerName)
	}
}
