// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package resolver

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"time"

	"github.com/miekg/dns"
	"github.com/quic-go/quic-go"
)

// DoQResolver implements DNS over QUIC (RFC 9250)
type DoQResolver struct {
	address   string
	port      string
	timeout   time.Duration
	tlsConfig *tls.Config
}

// NewDoQResolver creates a new DoQ resolver
func NewDoQResolver(address, port, serverName string, insecureSkipVerify bool, timeout time.Duration) *DoQResolver {
	tlsConfig := &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: insecureSkipVerify,
		NextProtos:         []string{"doq"},
	}

	return &DoQResolver{
		address:   address,
		port:      port,
		timeout:   timeout,
		tlsConfig: tlsConfig,
	}
}

// Query performs a DNS query using DoQ
func (r *DoQResolver) Query(ctx context.Context, hostname string, qtype uint16) QueryResult {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), qtype)

	wireMsg, err := msg.Pack()
	if err != nil {
		return QueryResult{Err: fmt.Errorf("failed to pack DNS message: %w", err)}
	}

	serverAddr := fmt.Sprintf("%s:%s", r.address, r.port)

	start := time.Now()

	queryCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	conn, err := quic.DialAddr(queryCtx, serverAddr, r.tlsConfig, &quic.Config{
		HandshakeIdleTimeout: r.timeout,
		MaxIdleTimeout:       r.timeout,
	})
	if err != nil {
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("QUIC dial failed: %w", err),
		}
	}
	defer conn.CloseWithError(0, "")

	stream, err := conn.OpenStreamSync(queryCtx)
	if err != nil {
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("failed to open QUIC stream: %w", err),
		}
	}

	// DoQ uses a 2-byte length prefix (RFC 9250)
	lengthPrefix := []byte{byte(len(wireMsg) >> 8), byte(len(wireMsg))}
	if _, err := stream.Write(lengthPrefix); err != nil {
		stream.Close()
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("failed to write length prefix: %w", err),
		}
	}
	if _, err := stream.Write(wireMsg); err != nil {
		stream.Close()
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("failed to write DNS message: %w", err),
		}
	}

	// Gracefully close send side (sends FIN) per RFC 9250
	if err := stream.Close(); err != nil {
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("failed to close send side: %w", err),
		}
	}

	// Read response length prefix
	respLengthBuf := make([]byte, 2)
	if _, err := io.ReadFull(stream, respLengthBuf); err != nil {
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("failed to read response length: %w", err),
		}
	}
	respLength := int(respLengthBuf[0])<<8 | int(respLengthBuf[1])

	// Read the full response
	respBuf := make([]byte, respLength)
	if _, err := io.ReadFull(stream, respBuf); err != nil {
		return QueryResult{
			Duration: time.Since(start),
			Err:      fmt.Errorf("failed to read response: %w", err),
		}
	}
	duration := time.Since(start)

	response := new(dns.Msg)
	if err := response.Unpack(respBuf); err != nil {
		return QueryResult{
			Duration: duration,
			Err:      fmt.Errorf("failed to unpack DNS response: %w", err),
		}
	}

	return QueryResult{
		Response: response,
		Duration: duration,
		Err:      nil,
	}
}

// Protocol returns the protocol identifier
func (r *DoQResolver) Protocol() string {
	return "doq"
}

// Close releases resources
func (r *DoQResolver) Close() error {
	return nil
}
