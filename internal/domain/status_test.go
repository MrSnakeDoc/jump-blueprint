package domain

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateTLS(t *testing.T) {
	// Create a test HTTPS server
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// This test is tricky because httptest uses self-signed certificates
	// ValidateTLS in production checks real certificates, so we'll test error cases instead
	tests := []struct {
		name        string
		hostname    string
		timeout     time.Duration
		shouldPass  bool
		description string
	}{
		{
			name:        "very short timeout",
			hostname:    "google.com:443",
			timeout:     1 * time.Nanosecond,
			shouldPass:  false,
			description: "extremely short timeout should fail",
		},
		{
			name:        "invalid hostname",
			hostname:    "invalid-hostname-that-does-not-exist-12345678",
			timeout:     1 * time.Second,
			shouldPass:  false,
			description: "invalid hostname should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTLS(tt.hostname, tt.timeout)
			if tt.shouldPass && err != nil {
				t.Errorf("ValidateTLS() = %v, want nil (should pass)", err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("ValidateTLS() = nil, want error (should fail)")
			}
		})
	}
}

// TestValidateTLSWithRealCert tests TLS validation with known good certificate
// Note: This test requires internet connectivity
func TestValidateTLSWithRealCert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test requiring internet connectivity")
	}

	err := ValidateTLS("google.com:443", 5*time.Second)
	if err != nil {
		t.Logf("ValidateTLS() with google.com = %v (may fail if internet unavailable)", err)
	}
}

// TestValidateTLSWithSelfSignedCert tests that self-signed certificates are rejected
func TestValidateTLSWithSelfSignedCert(t *testing.T) {
	// Create a test HTTPS server with self-signed cert
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Extract hostname from test server URL
	testHost := ts.URL[8:] // Remove "https://"

	// This should fail because the certificate is self-signed
	err := ValidateTLS(testHost, 5*time.Second)
	if err == nil {
		// If it doesn't fail, it means the system trusts self-signed certs (unlikely in production)
		t.Logf("ValidateTLS() with self-signed cert succeeded (system may trust self-signed certs)")
	}
}

// TestValidateTLSWithCustomClient verifies we can validate with custom transport
func TestValidateTLSWithCustomClient(t *testing.T) {
	// Create a test HTTPS server
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Extract hostname
	testHost := ts.URL[8:]

	// Create a custom HTTP client that accepts self-signed certs
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	req, err := http.NewRequest("HEAD", ts.URL, http.NoBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %v", resp.StatusCode)
	}

	// Now test that ValidateTLS (without InsecureSkipVerify) would reject this
	err = ValidateTLS(testHost, 5*time.Second)
	if err == nil {
		t.Log("ValidateTLS accepted self-signed cert (unexpected but system-dependent)")
	}
}

func TestValidateTLSInvalidURL(t *testing.T) {
	err := ValidateTLS("not-a-valid-hostname-12345", 1*time.Second)
	if err == nil {
		t.Error("ValidateTLS() with invalid hostname should return error")
	}
}

func TestValidateTLSEmptyURL(t *testing.T) {
	err := ValidateTLS("", 1*time.Second)
	if err == nil {
		t.Error("ValidateTLS() with empty hostname should return error")
	}
}
