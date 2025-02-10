package main

import (
	"crypto/rand"
	"encoding/base32"
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
func PerformDNSQuery(hostname, server string) (float64, error) {
	var duration float64
	var err error
	client := new(dns.Client)
	message := new(dns.Msg)

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
				duration, err = PerformDNSQuery(hostname, serverAddr)
				if err == nil { // successful lookup
					if config.VerboseLogging {
						log.Printf("Looking up %s from %s - success - %3.2f sec", hostname, serverAddr, duration)
					}
					dnsQuerySuccess.WithLabelValues(domain.Name, serverAddr).Inc()
				} else { // failed lookup
					if config.VerboseLogging {
						log.Printf("Looking up %s from %s - failed  - %3.2f sec (setting to 10 sec) - error: %s", hostname, serverAddr, duration, err)
					}
					dnsQueryFailures.WithLabelValues(domain.Name, serverAddr).Inc()
					duration = float64(time.Second * 10)
				}

				dnsQueryDuration.WithLabelValues(domain.Name, serverAddr).Observe(duration)

				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func main() {
	config, err := LoadConfig("config.yaml")
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
	log.Printf("Starting Prometheus metrics server on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
