package network

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// NewHTTPClient creates an HTTP client optimized for speed testing
func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			},
		},
		Timeout: 60 * time.Second,
	}
}

// DownloadClient creates an HTTP client optimized for downloads
// No timeout for large downloads
func DownloadClient() *http.Client {
	client := NewHTTPClient()
	client.Timeout = 0 // No timeout for large downloads
	return client
}

// UploadClient creates an HTTP client optimized for uploads
func UploadClient() *http.Client {
	return NewHTTPClient()
}
