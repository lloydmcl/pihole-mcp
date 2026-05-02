// Pi-hole MCP Server — enables AI assistants to manage Pi-hole v6 instances.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hexamatic/pihole-mcp/internal/config"
	"github.com/hexamatic/pihole-mcp/internal/pihole"
	piholeserver "github.com/hexamatic/pihole-mcp/internal/server"
	"github.com/hexamatic/pihole-mcp/internal/telemetry"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	version := flag.Bool("version", false, "Print version and exit")
	transport := flag.String("transport", "stdio", "Transport type: stdio, http, or sse")
	address := flag.String("address", "localhost:8080", "Listen address for http/sse transports")
	flag.Parse()

	if *version {
		fmt.Println("pihole-mcp " + piholeserver.Version)
		return
	}

	log.SetOutput(os.Stderr)

	if err := run(*transport, *address); err != nil {
		log.Fatal(err)
	}
}

func run(transport, address string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	client := pihole.New(cfg.URL, cfg.Password,
		pihole.WithTimeout(cfg.RequestTimeout))
	defer client.Close()

	srv := piholeserver.New(client)

	tp, err := telemetry.Init("pihole-mcp", piholeserver.Version)
	if err != nil {
		return fmt.Errorf("telemetry init error: %w", err)
	}
	if tp != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = tp.Shutdown(ctx)
		}()
	}

	switch transport {
	case "stdio":
		return server.ServeStdio(srv)

	case "http":
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		httpSrv := server.NewStreamableHTTPServer(srv)
		log.Printf("starting HTTP transport on %s", address)

		go func() {
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := httpSrv.Shutdown(shutdownCtx); err != nil {
				log.Printf("shutdown error: %v", err)
			}
		}()

		return httpSrv.Start(address)

	case "sse":
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		sseSrv := server.NewSSEServer(srv)
		log.Printf("starting SSE transport on %s", address)

		go func() {
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := sseSrv.Shutdown(shutdownCtx); err != nil {
				log.Printf("shutdown error: %v", err)
			}
		}()

		return sseSrv.Start(address)

	default:
		return fmt.Errorf("unknown transport: %s (expected stdio, http, or sse)", transport)
	}
}
