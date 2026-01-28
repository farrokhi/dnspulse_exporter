// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2026 Babak Farrokhi

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// TLSConfig holds TLS-specific configuration for encrypted protocols
type TLSConfig struct {
	ServerName         string `yaml:"server_name"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

// DNSServer represents a single DNS server configuration
type DNSServer struct {
	Address  string     `yaml:"address"`
	Port     string     `yaml:"port"`
	Protocol string     `yaml:"protocol"`
	TLS      *TLSConfig `yaml:"tls,omitempty"`
}

// Domain represents a domain to probe
type Domain struct {
	Name   string `yaml:"name"`
	Probes int    `yaml:"probes"`
}

// Config structure for YAML configuration file
type Config struct {
	Domains        []Domain    `yaml:"domains"`
	DNSServers     []DNSServer `yaml:"dns_servers"`
	ListenAddress  string      `yaml:"listen_addr"`
	ListenPort     string      `yaml:"listen_port"`
	VerboseLogging bool        `yaml:"verbose_logging"`
	Timeout        int64       `yaml:"timeout"`
}

// Supported DNS protocols
const (
	ProtocolDo53UDP = "do53-udp"
	ProtocolDo53TCP = "do53-tcp"
	ProtocolDoT     = "dot"
	ProtocolDoH     = "doh"
	ProtocolDoH3    = "doh3"
	ProtocolDoQ     = "doq"
)

// ValidProtocols lists all supported DNS protocols
var ValidProtocols = map[string]bool{
	ProtocolDo53UDP: true,
	ProtocolDo53TCP: true,
	ProtocolDoT:     true,
	ProtocolDoH:     true,
	ProtocolDoH3:    true,
	ProtocolDoQ:     true,
}

// IsEncryptedProtocol returns true if the protocol uses TLS/encryption
func IsEncryptedProtocol(protocol string) bool {
	return protocol == ProtocolDoT || protocol == ProtocolDoH ||
		protocol == ProtocolDoH3 || protocol == ProtocolDoQ
}

// Load reads YAML configuration from a file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	config.applyDefaults()

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// applyDefaults sets default values for optional fields
func (c *Config) applyDefaults() {
	for i := range c.DNSServers {
		if c.DNSServers[i].Protocol == "" {
			c.DNSServers[i].Protocol = ProtocolDo53UDP
		}
		if c.DNSServers[i].Port == "" {
			c.DNSServers[i].Port = defaultPortForProtocol(c.DNSServers[i].Protocol)
		}
	}
}

// validate checks the configuration for errors
func (c *Config) validate() error {
	for i, server := range c.DNSServers {
		if !ValidProtocols[server.Protocol] {
			return fmt.Errorf("invalid protocol '%s' for server %s", server.Protocol, server.Address)
		}

		if IsEncryptedProtocol(server.Protocol) {
			if server.TLS == nil {
				c.DNSServers[i].TLS = &TLSConfig{ServerName: server.Address}
			} else if server.TLS.ServerName == "" {
				c.DNSServers[i].TLS.ServerName = server.Address
			}
		}
	}
	return nil
}

// defaultPortForProtocol returns the standard port for each protocol
func defaultPortForProtocol(protocol string) string {
	switch protocol {
	case ProtocolDo53UDP, ProtocolDo53TCP:
		return "53"
	case ProtocolDoT, ProtocolDoQ:
		return "853"
	case ProtocolDoH, ProtocolDoH3:
		return "443"
	default:
		return "53"
	}
}
