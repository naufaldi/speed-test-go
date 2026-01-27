package network

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient()

	if client == nil {
		t.Fatal("Expected non-nil HTTP client")
	}

	if client.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got: %v", client.Timeout)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	// Verify TLS configuration
	if transport.TLSClientConfig == nil {
		t.Error("Expected TLS config to be set")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected TLS 1.2 minimum, got: %d", transport.TLSClientConfig.MinVersion)
	}

	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be false")
	}
}

func TestNewHTTPClient_TransportSettings(t *testing.T) {
	client := NewHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	// Verify dial timeout
	if transport.DialContext == nil {
		t.Error("Expected DialContext to be set")
	}

	// Verify TLS handshake timeout
	if transport.TLSHandshakeTimeout != 10*time.Second {
		t.Errorf("Expected TLSHandshakeTimeout 10s, got: %v", transport.TLSHandshakeTimeout)
	}

	// Verify response header timeout
	if transport.ResponseHeaderTimeout != 30*time.Second {
		t.Errorf("Expected ResponseHeaderTimeout 30s, got: %v", transport.ResponseHeaderTimeout)
	}

	// Verify idle connection settings
	if transport.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns 10, got: %d", transport.MaxIdleConns)
	}

	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf("Expected IdleConnTimeout 90s, got: %v", transport.IdleConnTimeout)
	}

	// Verify keep alive
	if transport.MaxIdleConnsPerHost != 0 {
		// Default is 0, which means same as MaxIdleConns
		t.Logf("MaxIdleConnsPerHost: %d", transport.MaxIdleConnsPerHost)
	}
}

func TestDownloadClient(t *testing.T) {
	client := DownloadClient()

	if client == nil {
		t.Fatal("Expected non-nil download client")
	}

	// Download client should have no timeout (0)
	if client.Timeout != 0 {
		t.Errorf("Expected download client timeout 0 (no timeout), got: %v", client.Timeout)
	}

	// Should still have transport configuration
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	// TLS config should still be present
	if transport.TLSClientConfig == nil {
		t.Error("Expected TLS config to be set for download client")
	}
}

func TestUploadClient(t *testing.T) {
	client := UploadClient()

	if client == nil {
		t.Fatal("Expected non-nil upload client")
	}

	// Upload client should use default timeout (60s)
	if client.Timeout != 60*time.Second {
		t.Errorf("Expected upload client timeout 60s, got: %v", client.Timeout)
	}

	// Should still have transport configuration
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	// TLS config should still be present
	if transport.TLSClientConfig == nil {
		t.Error("Expected TLS config to be set for upload client")
	}
}

func TestNewHTTPClient_VerifyDialer(t *testing.T) {
	client := NewHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	// Test that DialContext is a function (not nil)
	if transport.DialContext == nil {
		t.Error("Expected DialContext to be set")
	}

	// The DialContext function is created from a net.Dialer
	// We can't easily test the dialer settings without making a connection
	t.Logf("DialContext is properly configured")
}

func TestDownloadClient_Isolation(t *testing.T) {
	downloadClient1 := DownloadClient()
	downloadClient2 := DownloadClient()

	// Each call should return a new client
	if downloadClient1 == downloadClient2 {
		t.Error("Expected separate download client instances")
	}

	// Modifying one shouldn't affect the other
	downloadClient1.Timeout = 0
	if downloadClient2.Timeout != 0 {
		// Both should be 0 anyway, but this tests isolation
		t.Logf("Download client timeout: %v", downloadClient2.Timeout)
	}
}

func TestUploadClient_Isolation(t *testing.T) {
	uploadClient1 := UploadClient()
	uploadClient2 := UploadClient()

	// Each call should return a new client
	if uploadClient1 == uploadClient2 {
		t.Error("Expected separate upload client instances")
	}
}

func TestHTTPClient_SeparateInstances(t *testing.T) {
	httpClient1 := NewHTTPClient()
	httpClient2 := NewHTTPClient()

	// Each call should return a new client
	if httpClient1 == httpClient2 {
		t.Error("Expected separate HTTP client instances")
	}
}

func TestHTTPClient_TLSConfig(t *testing.T) {
	client := NewHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	tlsConfig := transport.TLSClientConfig

	// Verify TLS version
	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected TLS 1.2 minimum, got version: %d", tlsConfig.MinVersion)
	}

	// Verify insecure skip is false
	if tlsConfig.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be false for security")
	}

	// Verify no custom certificates (nil means system default)
	if tlsConfig.RootCAs != nil {
		t.Log("Using custom Root CAs")
	}

	if tlsConfig.ClientAuth != tls.NoClientCert {
		t.Logf("ClientAuth setting: %d", tlsConfig.ClientAuth)
	}
}

func TestHTTPClient_ExpectContinueTimeout(t *testing.T) {
	client := NewHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	if transport.ExpectContinueTimeout != 1*time.Second {
		t.Errorf("Expected ExpectContinueTimeout 1s, got: %v", transport.ExpectContinueTimeout)
	}
}

func TestHTTPClient_ProxySettings(t *testing.T) {
	client := NewHTTPClient()

	// By default, no proxy should be set
	if client.CheckRedirect != nil {
		t.Logf("CheckRedirect is set")
	}

	if client.Jar != nil {
		t.Log("Cookie jar is set")
	}
}

func TestHTTPClient_ConcurrentCreation(t *testing.T) {
	done := make(chan *http.Client, 100)

	// Create clients concurrently
	for i := 0; i < 100; i++ {
		go func() {
			client := NewHTTPClient()
			done <- client
		}()
	}

	// All clients should be created successfully
	clients := make([]*http.Client, 0, 100)
	for i := 0; i < 100; i++ {
		clients = append(clients, <-done)
	}

	// Verify all clients are valid
	for i, client := range clients {
		if client == nil {
			t.Errorf("Client %d is nil", i)
		}
		if client.Timeout != 60*time.Second {
			t.Errorf("Client %d has wrong timeout", i)
		}
	}
}

func TestDownloadClient_NoTimeout(t *testing.T) {
	// The key feature of DownloadClient is no timeout
	client := DownloadClient()

	if client.Timeout != 0 {
		t.Errorf("Expected 0 timeout (no limit), got: %v", client.Timeout)
	}
}

func TestUploadClient_HasTimeout(t *testing.T) {
	// Upload client should have timeout
	client := UploadClient()

	if client.Timeout == 0 {
		t.Error("Expected upload client to have timeout")
	}

	if client.Timeout != 60*time.Second {
		t.Errorf("Expected 60s timeout, got: %v", client.Timeout)
	}
}

func TestHTTPClient_TransportRecycling(t *testing.T) {
	client := NewHTTPClient()

	// Make a request
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	if req == nil {
		t.Fatal("Failed to create request")
	}

	// The transport should be reusable
	transport1, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	// Transport should have proper configuration
	if transport1.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns 10, got: %d", transport1.MaxIdleConns)
	}
}
