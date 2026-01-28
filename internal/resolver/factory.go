// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package resolver

import (
	"fmt"
	"time"

	"dnspulse_exporter/internal/config"
)

// NewResolver creates a resolver based on the server configuration
func NewResolver(server config.DNSServer, timeout time.Duration) (Resolver, error) {
	serverName, insecure := extractTLSConfig(server)

	switch server.Protocol {
	case config.ProtocolDo53UDP:
		return NewDo53Resolver(server.Address, server.Port, false, timeout), nil
	case config.ProtocolDo53TCP:
		return NewDo53Resolver(server.Address, server.Port, true, timeout), nil
	case config.ProtocolDoT:
		return NewDoTResolver(server.Address, server.Port, serverName, insecure, timeout), nil
	case config.ProtocolDoH:
		return NewDoHResolver(server.Address, server.Port, serverName, insecure, timeout), nil
	case config.ProtocolDoH3:
		return NewDoH3Resolver(server.Address, server.Port, serverName, insecure, timeout), nil
	case config.ProtocolDoQ:
		return NewDoQResolver(server.Address, server.Port, serverName, insecure, timeout), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", server.Protocol)
	}
}

// extractTLSConfig extracts TLS settings from server config
func extractTLSConfig(server config.DNSServer) (serverName string, insecure bool) {
	serverName = server.Address
	if server.TLS != nil {
		if server.TLS.ServerName != "" {
			serverName = server.TLS.ServerName
		}
		insecure = server.TLS.InsecureSkipVerify
	}
	return
}
