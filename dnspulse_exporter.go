// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2025 Babak Farrokhi

package main

import (
	"crypto/rand"
	"encoding/base32"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

var (
	version   = "1.1"
	buildTime = "unknown"
)

// Config structure for YAML configuration file
type Config struct {
	Domains []struct {
		Name   string `yaml:"name"`
		Probes int    `yaml:"probes"`
	} `yaml:"domains"`
	DNSServers []struct {
		Address string `yaml:"address"`
		Port    string `yaml:"port"`
	} `yaml:"dns_servers"`
	ListenAddress  string `yaml:"listen_addr"`
	ListenPort     string `yaml:"listen_port"`
	VerboseLogging bool   `yaml:"verbose_logging"`
	Timeout        int64  `yaml:"timeout"`
}

var (
	dnsQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_query_duration_seconds",
			Help:    "Duration of DNS queries",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"domain", "server"},
	)
	dnsQuerySuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_query_success_total",
			Help: "Total successful DNS queries",
		},
		[]string{"domain", "server"},
	)
	dnsQueryFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_query_failures_total",
			Help: "Total failed DNS queries",
		},
		[]string{"domain", "server"},
	)
)

func init() {
	prometheus.MustRegister(dnsQueryDuration, dnsQuerySuccess, dnsQueryFailures)
}

// LoadConfig reads YAML configuration from a file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GenerateRandomPrefix creates a short random string to use as a hostname prefix
func GenerateRandomPrefix(len uint) string {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("Error generating random prefix:", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}

// PerformDNSQuery performs a DNS A record lookup for a given hostname
func PerformDNSQuery(hostname, server string, timeout int64) (float64, error) {
	var duration float64
	var err error
	client := new(dns.Client)
	message := new(dns.Msg)

	if timeout > 0 { // default timeout of 2000 milliseconds in implied
		client.Timeout = time.Duration(timeout) * time.Millisecond
	}
	message.SetQuestion(dns.Fqdn(hostname), dns.TypeA)

	// The call to `Exchange()` does not retry, nor falls back to TCP. This is intended as we only
	// care about success and timeout. You need to pick up a hostname that has a short answer that
	// fits in UDP response.
	start := time.Now()
	_, _, err = client.Exchange(message, server)
	duration = time.Since(start).Seconds()

	return duration, err
}

func runDNSQueries(config *Config) {
	var duration float64
	var err error

	for _, domain := range config.Domains {
		for _, server := range config.DNSServers {
			serverAddr := fmt.Sprintf("%s:%s", server.Address, server.Port)
			for i := 0; i < domain.Probes; i++ {
				prefix := GenerateRandomPrefix(5)
				hostname := fmt.Sprintf("%s.%s", prefix, domain.Name)
				duration, err = PerformDNSQuery(hostname, serverAddr, config.Timeout)
				if err == nil { // successful lookup
					if config.VerboseLogging {
						log.Printf("(%-25s)?(%s) - success - %-5.0f msec", hostname, serverAddr, duration*1000)
					}
					dnsQuerySuccess.WithLabelValues(domain.Name, serverAddr).Inc()
				} else { // failed lookup
					if config.VerboseLogging {
						log.Printf("(%-25s)?(%s) - failed  - %-5.0f msec - error: %s", hostname, serverAddr, duration*1000, err)
					}
					dnsQueryFailures.WithLabelValues(domain.Name, serverAddr).Inc()
				}

				dnsQueryDuration.WithLabelValues(domain.Name, serverAddr).Observe(duration)

				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func main() {

	var configFile string = "/etc/dnspulse.yml"
	var showVersion bool

	//flags.StringVar(&config.ListenAddress, "listen-addr", "0.0.0.0", "Listen address")
	//flags.StringVar(&config.ListenPort, "listen-port", "53", "Listen port")
	flag.StringVar(&configFile, "f", configFile, "Path to config file")
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Printf("dnspluse_exporter %s\n", version)
		os.Exit(0)
	}

	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	listenAddr := config.ListenAddress
	if listenAddr == "*" {
		listenAddr = ""
	}
	serverAddr := fmt.Sprintf("%s:%s", listenAddr, config.ListenPort)

	go func() {
		for {
			runDNSQueries(config)
			time.Sleep(30 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Starting Prometheus metrics server on %s", serverAddr)
	log.Fatal(server.ListenAndServe())
}
