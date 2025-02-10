package main

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	ListenAddress string `yaml:"listen_addr"`
	ListenPort    string `yaml:"listen_port"`
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
	data, err := ioutil.ReadFile(filename)
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
func GenerateRandomPrefix() string {
	b := make([]byte, 5)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("Error generating random prefix:", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}

// PerformDNSQuery performs a DNS A record lookup for a given hostname
func PerformDNSQuery(domainName, server string) ([]string, error) {
	var results []string
	var duration float64
	client := new(dns.Client)
	message := new(dns.Msg)

	prefix := GenerateRandomPrefix()
	hostname := fmt.Sprintf("%s.%s", prefix, domainName)

	message.SetQuestion(dns.Fqdn(hostname), dns.TypeA)

	start := time.Now()

	// This call does not retry, nor falls back to TCP. This is intended. You need to pick up a
	// hostname that has a short answer that fits in UDP response.
	response, _, err := client.Exchange(message, server)
	if err != nil {
		duration = float64(time.Second * 10)
	} else {
		duration = time.Since(start).Seconds()
	}

	dnsQueryDuration.WithLabelValues(domainName, server).Observe(duration)

	if err != nil {
		dnsQueryFailures.WithLabelValues(domainName, server).Inc()
		return nil, err
	}
	dnsQuerySuccess.WithLabelValues(domainName, server).Inc()

	for _, ans := range response.Answer {
		if a, ok := ans.(*dns.A); ok {
			results = append(results, a.A.String())
		}
	}

	return results, nil
}

func runDNSQueries(config *Config) {
	for _, domain := range config.Domains {
		for _, server := range config.DNSServers {
			serverAddr := fmt.Sprintf("%s:%s", server.Address, server.Port)
			for i := 0; i < domain.Probes; i++ {
				_, _ = PerformDNSQuery(domain.Name, serverAddr)
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
