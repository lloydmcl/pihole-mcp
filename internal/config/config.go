// Package config loads Pi-hole MCP server configuration from environment variables.
package config

import (
	"fmt"
	"net/url"
	"os"
	"time"
)

const (
	defaultRequestTimeout = 30 * time.Second
)

// Config holds the Pi-hole MCP server configuration.
type Config struct {
	// URL is the base URL of the Pi-hole instance (e.g. "http://192.168.1.2").
	URL string

	// Password is the admin password or application password for the Pi-hole API.
	Password string

	// RequestTimeout is the HTTP request timeout for Pi-hole API calls.
	RequestTimeout time.Duration
}

// Load reads configuration from environment variables and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		URL:            os.Getenv("PIHOLE_URL"),
		Password:       os.Getenv("PIHOLE_PASSWORD"),
		RequestTimeout: defaultRequestTimeout,
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("PIHOLE_URL environment variable is required")
	}

	if _, err := url.Parse(cfg.URL); err != nil {
		return nil, fmt.Errorf("PIHOLE_URL is not a valid URL: %w", err)
	}

	if _, ok := os.LookupEnv("PIHOLE_PASSWORD"); !ok {
		return nil, fmt.Errorf("PIHOLE_PASSWORD environment variable is required (set to empty string for no-password Pi-hole instances)")
	}

	if v := os.Getenv("PIHOLE_REQUEST_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("PIHOLE_REQUEST_TIMEOUT is not a valid duration: %w", err)
		}
		cfg.RequestTimeout = d
	}

	return cfg, nil
}
