package domain

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

// ValidateTLS checks if a service is reachable and has a valid TLS certificate
func ValidateTLS(hostname string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Construct HTTPS URL
	url := fmt.Sprintf("https://%s", hostname)

	// Create HTTP client with custom transport (short timeout)
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{
					Timeout:   timeout,
					KeepAlive: 0,
				}).DialContext(ctx, network, addr)
			},
			TLSHandshakeTimeout: timeout,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			DisableKeepAlives: true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects
			return http.ErrUseLastResponse
		},
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate TLS: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore close errors in validation context
	}()

	// Any response means the service is reachable with valid TLS
	return nil
}

// IsServiceHealthy checks if a service is healthy
func IsServiceHealthy(service *Service, timeout time.Duration) bool {
	if service == nil {
		return false
	}
	return ValidateTLS(service.Hostname, timeout) == nil
}

// ValidateMultiple validates multiple services and returns healthy ones
func ValidateMultiple(candidates []*Candidate, timeout time.Duration) *Candidate {
	for _, candidate := range candidates {
		if err := ValidateTLS(candidate.Service.Hostname, timeout); err == nil {
			// First healthy service wins
			return candidate
		}
	}
	return nil
}
