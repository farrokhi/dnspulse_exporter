// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2026 Babak Farrokhi

package prober

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"

	"dnspulse_exporter/internal/config"
	"dnspulse_exporter/internal/metrics"
	"dnspulse_exporter/internal/resolver"
)

// Prober orchestrates DNS queries across multiple resolvers
type Prober struct {
	config    *config.Config
	resolvers map[string]resolver.Resolver
	verbose   bool
}

// New creates a new Prober with resolvers for all configured servers
func New(cfg *config.Config) (*Prober, error) {
	timeout := time.Duration(cfg.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	resolvers := make(map[string]resolver.Resolver)
	for _, server := range cfg.DNSServers {
		key := serverKey(server)
		r, err := resolver.NewResolver(server, timeout)
		if err != nil {
			return nil, fmt.Errorf("failed to create resolver for %s: %w", server.Address, err)
		}
		resolvers[key] = r
	}

	return &Prober{
		config:    cfg,
		resolvers: resolvers,
		verbose:   cfg.VerboseLogging,
	}, nil
}

// serverKey generates a unique key for a server configuration
func serverKey(server config.DNSServer) string {
	return fmt.Sprintf("%s:%s:%s", server.Address, server.Port, server.Protocol)
}

// Run executes one round of DNS probes for all configured domains and servers
func (p *Prober) Run(ctx context.Context) {
	for _, domain := range p.config.Domains {
		for _, server := range p.config.DNSServers {
			key := serverKey(server)
			r := p.resolvers[key]

			serverAddr := fmt.Sprintf("%s:%s", server.Address, server.Port)
			protocol := r.Protocol()

			for i := 0; i < domain.Probes; i++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				prefix := generateRandomPrefix(5)
				hostname := fmt.Sprintf("%s.%s", prefix, domain.Name)

				result := r.Query(ctx, hostname, dns.TypeA)
				duration := result.Duration.Seconds()
				success := result.Err == nil

				if p.verbose {
					if success {
						log.Printf("[%s] (%-25s)?(%s) - success - %-5.0f msec",
							protocol, hostname, serverAddr, duration*1000)
					} else {
						log.Printf("[%s] (%-25s)?(%s) - failed  - %-5.0f msec - error: %s",
							protocol, hostname, serverAddr, duration*1000, result.Err)
					}
				}

				metrics.RecordQuery(domain.Name, serverAddr, protocol, duration, success)

				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

// Close releases all resolver resources
func (p *Prober) Close() {
	for name, r := range p.resolvers {
		if err := r.Close(); err != nil {
			log.Printf("warning: failed to close resolver %s: %v", name, err)
		}
	}
}

// generateRandomPrefix creates a short random string to use as a hostname prefix
func generateRandomPrefix(length uint) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Warning: error generating random prefix: %v", err)
		return "random"
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}
