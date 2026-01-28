// SPDX-License-Identifier: BSD-2-Clause
// Copyright (c) 2026 Babak Farrokhi

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"dnspulse_exporter/internal/config"
	"dnspulse_exporter/internal/prober"
)

var (
	version   = "1.1"
	gitCommit = "dev"
	buildTime = "unknown"
)

var configFile string

func main() {
	rootCmd := &cobra.Command{
		Use:   "dnspulse_exporter",
		Short: "Prometheus exporter for DNS query metrics",
		Run:   run,
	}

	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, gitCommit, buildTime)
	rootCmd.Flags().StringVarP(&configFile, "config", "f", "/etc/dnspulse.yml", "path to config file")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	p, err := prober.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create prober: %v", err)
	}
	defer p.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				p.Run(ctx)
				time.Sleep(30 * time.Second)
			}
		}
	}()

	listenAddr := cfg.ListenAddress
	if listenAddr == "*" {
		listenAddr = ""
	}
	serverAddr := fmt.Sprintf("%s:%s", listenAddr, cfg.ListenPort)

	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Starting Prometheus metrics server on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
