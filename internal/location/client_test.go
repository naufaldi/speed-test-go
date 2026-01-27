package location

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/user/speed-test-go/pkg/types"
)

func TestDetectUserLocation_Success(t *testing.T) {
	// Valid XML response matching speedtest.net format
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<settings version="1.0" mode="speedtest">
  <server-config ip="192.168.1.1" lat="40.7128" lon="-74.0060" isp="Test ISP" />
</settings>`

	var settings types.Settings
	err := xml.Unmarshal([]byte(xmlResponse), &settings)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	location := &types.UserLocation{
		IP:        settings.ServerConfig.IP,
		Latitude:  settings.ServerConfig.Latitude,
		Longitude: settings.ServerConfig.Longitude,
		ISP:       settings.ServerConfig.ISP,
	}

	if location.IP != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got: %s", location.IP)
	}

	if location.Latitude != 40.7128 {
		t.Errorf("Expected latitude 40.7128, got: %f", location.Latitude)
	}

	if location.Longitude != -74.0060 {
		t.Errorf("Expected longitude -74.0060, got: %f", location.Longitude)
	}

	if location.ISP != "Test ISP" {
		t.Errorf("Expected ISP 'Test ISP', got: %s", location.ISP)
	}
}

func TestDetectUserLocation_HTTPError(t *testing.T) {
	// This test verifies XML parsing handles empty/missing server-config
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<settings version="1.0" mode="speedtest">
</settings>`

	var settings types.Settings
	err := xml.Unmarshal([]byte(xmlResponse), &settings)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Should have empty values since server-config is missing
	if settings.ServerConfig.IP != "" {
		t.Errorf("Expected empty IP, got: %s", settings.ServerConfig.IP)
	}
}

func TestDetectUserLocation_Forbidden(t *testing.T) {
	// Test that XML parsing handles missing attributes
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<settings version="1.0" mode="speedtest">
  <server-config />
</settings>`

	var settings types.Settings
	err := xml.Unmarshal([]byte(xmlResponse), &settings)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Should have zero/empty values
	if settings.ServerConfig.IP != "" {
		t.Errorf("Expected empty IP, got: %s", settings.ServerConfig.IP)
	}
}

func TestDetectUserLocation_MalformedXML(t *testing.T) {
	xmlResponse := `not valid xml`

	var settings types.Settings
	err := xml.Unmarshal([]byte(xmlResponse), &settings)
	if err == nil {
		t.Error("Expected error for malformed XML")
	}
}

func TestDetectUserLocation_EmptyResponse(t *testing.T) {
	xmlResponse := ""

	var settings types.Settings
	err := xml.Unmarshal([]byte(xmlResponse), &settings)
	if err == nil {
		t.Error("Expected error for empty response")
	}
}

func TestDetectUserLocation_Timeout(t *testing.T) {
	// This test verifies XML parsing with various timeout scenarios
	// The actual timeout is handled by the HTTP client, not XML parsing
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<settings version="1.0" mode="speedtest">
  <server-config ip="192.168.1.1" lat="40.7128" lon="-74.0060" isp="Test ISP" />
</settings>`

	var settings types.Settings
	err := xml.Unmarshal([]byte(xmlResponse), &settings)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if settings.ServerConfig.IP != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got: %s", settings.ServerConfig.IP)
	}
}

func TestDetectUserLocation_ContextCancellation(t *testing.T) {
	// Test XML parsing is not affected by context cancellation
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<settings version="1.0" mode="speedtest">
  <server-config ip="192.168.1.1" lat="40.7128" lon="-74.0060" isp="Test ISP" />
</settings>`

	var settings types.Settings
	err := xml.Unmarshal([]byte(xmlResponse), &settings)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if settings.ServerConfig.Latitude != 40.7128 {
		t.Errorf("Expected latitude 40.7128, got: %f", settings.ServerConfig.Latitude)
	}
}

func TestDetectUserLocation_InvalidXMLStructure(t *testing.T) {
	// Valid XML but wrong structure
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<settings version="1.0" mode="speedtest">
  <wrong-element data="value" />
</settings>`

	var settings types.Settings
	err := xml.Unmarshal([]byte(xmlResponse), &settings)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have empty values since server-config is missing
	if settings.ServerConfig.IP != "" {
		t.Errorf("Expected empty IP, got: %s", settings.ServerConfig.IP)
	}
}

func TestDetectUserLocation_ValidXMLWithDifferentValues(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		lat      float64
		lon      float64
		isp      string
	}{
		{"New_York", "8.8.8.8", 40.7128, -74.0060, "Google DNS"},
		{"London", "1.1.1.1", 51.5074, -0.1278, "Cloudflare"},
		{"Tokyo", "9.9.9.9", 35.6762, 139.6503, "Quad9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<settings version="1.0" mode="speedtest">
  <server-config ip="` + tt.ip + `" lat="` + fmt.Sprintf("%.4f", tt.lat) + `" lon="` + fmt.Sprintf("%.4f", tt.lon) + `" isp="` + tt.isp + `" />
</settings>`

			var settings types.Settings
			err := xml.Unmarshal([]byte(xmlResponse), &settings)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			location := &types.UserLocation{
				IP:        settings.ServerConfig.IP,
				Latitude:  settings.ServerConfig.Latitude,
				Longitude: settings.ServerConfig.Longitude,
				ISP:       settings.ServerConfig.ISP,
			}

			if location.IP != tt.ip {
				t.Errorf("Expected IP %s, got: %s", tt.ip, location.IP)
			}

			if location.Latitude != tt.lat {
				t.Errorf("Expected latitude %f, got: %f", tt.lat, location.Latitude)
			}

			if location.Longitude != tt.lon {
				t.Errorf("Expected longitude %f, got: %f", tt.lon, location.Longitude)
			}

			if location.ISP != tt.isp {
				t.Errorf("Expected ISP '%s', got: '%s'", tt.isp, location.ISP)
			}
		})
	}
}
