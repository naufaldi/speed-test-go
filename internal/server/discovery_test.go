package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/speed-test-go/pkg/types"
)

func TestFetchServerList_Success(t *testing.T) {
	// Valid XML response matching speedtest.net format
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<servers version="1.0" mode="speedtest">
  <server url="http://test1.example.com/speedtest/upload.php" lat="40.7128" lon="-74.0060" name="Test Server 1" country="USA" sponsor="Test Sponsor 1" id="1" host="test1.example.com" />
  <server url="http://test2.example.com/speedtest/upload.php" lat="34.0522" lon="-118.2437" name="Test Server 2" country="USA" sponsor="Test Sponsor 2" id="2" host="test2.example.com" />
</servers>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/js/servers" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(xmlResponse))
	}))
	defer server.Close()

	ctx := context.Background()
	testURL := server.URL + "/api/js/servers?engine=js&limit=10"

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}
}

func TestFetchServerList_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	testURL := server.URL + "/api/js/servers?engine=js&limit=10"

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got: %d", resp.StatusCode)
	}
}

func TestFetchServerList_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	ctx := context.Background()
	testURL := server.URL + "/api/js/servers?engine=js&limit=10"

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got: %d", resp.StatusCode)
	}
}

func TestFetchServerList_EmptyServerList(t *testing.T) {
	// XML with no servers
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<servers version="1.0" mode="speedtest">
</servers>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(xmlResponse))
	}))
	defer server.Close()

	ctx := context.Background()
	testURL := server.URL + "/api/js/servers?engine=js&limit=10"

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}
}

func TestFetchServerList_SingleServer(t *testing.T) {
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<servers version="1.0" mode="speedtest">
  <server url="http://single.example.com/speedtest/upload.php" lat="40.7128" lon="-74.0060" name="Single Server" country="USA" sponsor="Single Sponsor" id="999" host="single.example.com" />
</servers>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(xmlResponse))
	}))
	defer server.Close()

	ctx := context.Background()
	testURL := server.URL + "/api/js/servers?engine=js&limit=10"

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}
}

func TestGetServerHost_HTTPProtocol(t *testing.T) {
	server := &types.Server{
		URL: "http://example.com/speedtest/upload.php",
	}

	host := GetServerHost(server)

	if host != "example.com" {
		t.Errorf("Expected 'example.com', got: %s", host)
	}
}

func TestGetServerHost_HTTPSProtocol(t *testing.T) {
	server := &types.Server{
		URL: "https://secure.example.com/speedtest/upload.php",
	}

	host := GetServerHost(server)

	if host != "secure.example.com" {
		t.Errorf("Expected 'secure.example.com', got: %s", host)
	}
}

func TestGetServerHost_WithPath(t *testing.T) {
	server := &types.Server{
		URL: "http://example.com/deep/path/to/upload.php",
	}

	host := GetServerHost(server)

	if host != "example.com" {
		t.Errorf("Expected 'example.com', got: %s", host)
	}
}

func TestGetServerHost_NoProtocol(t *testing.T) {
	server := &types.Server{
		URL: "example.com/upload.php",
	}

	host := GetServerHost(server)

	// Should strip path even without protocol
	if host != "example.com" {
		t.Errorf("Expected 'example.com', got: %s", host)
	}
}

func TestGetServerHost_IPAddress(t *testing.T) {
	server := &types.Server{
		URL: "http://192.168.1.1:8080/speedtest/upload.php",
	}

	host := GetServerHost(server)

	if host != "192.168.1.1:8080" {
		t.Errorf("Expected '192.168.1.1:8080', got: %s", host)
	}
}

func TestGetServerHost_QueryString(t *testing.T) {
	server := &types.Server{
		URL: "http://example.com/upload.php?key=value",
	}

	host := GetServerHost(server)

	if host != "example.com" {
		t.Errorf("Expected 'example.com', got: %s", host)
	}
}
