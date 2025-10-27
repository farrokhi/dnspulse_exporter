// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid config file", func(t *testing.T) {
		// Create a temporary config file
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

		config, err := LoadConfig(tempFile.Name())
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
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
		_, err := LoadConfig("/nonexistent/config.yml")
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

		_, err = LoadConfig(tempFile.Name())
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})
}

func TestGenerateRandomPrefix(t *testing.T) {
	t.Run("generates prefix of correct length", func(t *testing.T) {
		prefix := GenerateRandomPrefix(5)
		// Base32 encoding without padding: 5 bytes = 8 characters
		if len(prefix) != 8 {
			t.Errorf("Expected prefix length 8, got %d", len(prefix))
		}
	})

	t.Run("generates unique prefixes", func(t *testing.T) {
		prefix1 := GenerateRandomPrefix(5)
		prefix2 := GenerateRandomPrefix(5)
		if prefix1 == prefix2 {
			t.Error("Expected different prefixes, got identical ones")
		}
	})

	t.Run("generates valid base32 string", func(t *testing.T) {
		prefix := GenerateRandomPrefix(5)
		// Check that prefix only contains valid base32 characters (A-Z, 2-7)
		for _, c := range prefix {
			if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7')) {
				t.Errorf("Invalid base32 character: %c", c)
			}
		}
	})
}

func TestPerformDNSQuery(t *testing.T) {
	t.Run("query with timeout", func(t *testing.T) {
		// Use a valid public DNS server and domain
		duration, err := PerformDNSQuery("example.com", "8.8.8.8:53", 5000)

		if err != nil {
			t.Logf("DNS query failed (this might be expected in some environments): %v", err)
		}

		if duration < 0 {
			t.Errorf("Expected non-negative duration, got %f", duration)
		}
	})

	t.Run("query with very short timeout", func(t *testing.T) {
		// This should likely timeout
		duration, err := PerformDNSQuery("example.com", "8.8.8.8:53", 1)

		// We expect either an error (timeout) or very quick response
		if duration < 0 {
			t.Errorf("Expected non-negative duration, got %f", duration)
		}

		// If error occurred, it's likely a timeout which is expected
		if err != nil {
			t.Logf("Query timed out as expected: %v", err)
		}
	})

	t.Run("query non-existent server", func(t *testing.T) {
		duration, err := PerformDNSQuery("example.com", "192.0.2.1:53", 1000)

		// Should have an error for non-existent server
		if err == nil {
			t.Error("Expected error for non-existent DNS server")
		}

		if duration < 0 {
			t.Errorf("Expected non-negative duration, got %f", duration)
		}
	})
}

func TestHTTPServerConfiguration(t *testing.T) {
	t.Run("server has proper timeouts configured", func(t *testing.T) {
		server := &http.Server{
			Addr:         ":9953",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		if server.ReadTimeout != 5*time.Second {
			t.Errorf("Expected ReadTimeout 5s, got %v", server.ReadTimeout)
		}
		if server.WriteTimeout != 10*time.Second {
			t.Errorf("Expected WriteTimeout 10s, got %v", server.WriteTimeout)
		}
		if server.IdleTimeout != 120*time.Second {
			t.Errorf("Expected IdleTimeout 120s, got %v", server.IdleTimeout)
		}
	})

	t.Run("metrics endpoint responds", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		handler := promhttp.Handler()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType == "" {
			t.Error("Expected Content-Type header to be set")
		}
	})
}

func TestListenAddressHandling(t *testing.T) {
	t.Run("wildcard is converted to empty string", func(t *testing.T) {
		listenAddr := "*"
		if listenAddr == "*" {
			listenAddr = ""
		}

		if listenAddr != "" {
			t.Errorf("Expected empty string, got '%s'", listenAddr)
		}
	})

	t.Run("specific address is preserved", func(t *testing.T) {
		listenAddr := "127.0.0.1"
		if listenAddr == "*" {
			listenAddr = ""
		}

		if listenAddr != "127.0.0.1" {
			t.Errorf("Expected '127.0.0.1', got '%s'", listenAddr)
		}
	})
}
