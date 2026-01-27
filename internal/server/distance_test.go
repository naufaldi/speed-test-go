package server

import (
	"math"
	"testing"

	"github.com/user/speed-test-go/pkg/types"
)

func TestCalculateDistance_SameLocation(t *testing.T) {
	// Same coordinates should result in distance of 0
	lat := 40.7128
	lon := -74.0060

	distance := CalculateDistance(lat, lon, lat, lon)

	if distance != 0 {
		t.Errorf("Expected distance 0 for same location, got: %f", distance)
	}
}

func TestCalculateDistance_KnownDistance(t *testing.T) {
	// Test with known distances for validation
	testCases := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
		delta    float64
	}{
		{
			name:     "New York to Los Angeles",
			lat1:     40.7128,
			lon1:     -74.0060,
			lat2:     34.0522,
			lon2:     -118.2437,
			expected: 3940, // Approximately 3940 km
			delta:    50,   // Allow 50km tolerance
		},
		{
			name:     "London to Paris",
			lat1:     51.5074,
			lon1:     -0.1278,
			lat2:     48.8566,
			lon2:     2.3522,
			expected: 344, // Approximately 344 km
			delta:    10,
		},
		{
			name:     "Tokyo to Sydney",
			lat1:     35.6762,
			lon1:     139.6503,
			lat2:     -33.8688,
			lon2:     151.2093,
			expected: 7820, // Approximately 7820 km
			delta:    100,
		},
		{
			name:     "Equator points",
			lat1:     0,
			lon1:     0,
			lat2:     0,
			lon2:     10,   // 10 degrees apart on equator
			expected: 1112, // Approximately 1112 km (111 km per degree)
			delta:    10,
		},
		{
			name:     "Same longitude, different latitude",
			lat1:     0,
			lon1:     0,
			lat2:     10,
			lon2:     0,
			expected: 1112, // Approximately 1112 km
			delta:    10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			distance := CalculateDistance(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
			diff := math.Abs(distance - tc.expected)
			if diff > tc.delta {
				t.Errorf("Expected distance ~%f km, got: %f km (diff: %f km)", tc.expected, distance, diff)
			}
		})
	}
}

func TestCalculateDistance_Poles(t *testing.T) {
	// Test edge cases near poles
	testCases := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
		delta    float64
	}{
		{
			name:     "North Pole to Equator",
			lat1:     90,
			lon1:     0,
			lat2:     0,
			lon2:     0,
			expected: 10007, // Approximately 10007 km (quarter of Earth's circumference)
			delta:    50,
		},
		{
			name:     "South Pole to Equator",
			lat1:     -90,
			lon1:     0,
			lat2:     0,
			lon2:     0,
			expected: 10007,
			delta:    50,
		},
		{
			name:     "North to South Pole",
			lat1:     90,
			lon1:     0,
			lat2:     -90,
			lon2:     0,
			expected: 20015, // Half of Earth's circumference
			delta:    100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			distance := CalculateDistance(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
			diff := math.Abs(distance - tc.expected)
			if diff > tc.delta {
				t.Errorf("Expected distance ~%f km, got: %f km (diff: %f km)", tc.expected, distance, diff)
			}
		})
	}
}

func TestCalculateDistance_AntipodalPoints(t *testing.T) {
	// Antipodal points are exactly opposite on the sphere
	// Use exact antipodal coordinates for New York
	lat1 := 40.7128
	lon1 := -74.0060
	lat2 := -40.7128           // Exactly opposite latitude
	lon2 := 105.9940           // Approximately opposite longitude (lon1 + 180)

	distance := CalculateDistance(lat1, lon1, lat2, lon2)

	// Antipodal distance should be approximately half Earth's circumference
	expected := 20015.0 // Half of Earth's circumference in km
	delta := 500.0      // Allow some tolerance for the longitude not being exact

	diff := math.Abs(distance - expected)
	if diff > delta {
		t.Errorf("Expected ~%f km for antipodal points, got: %f km", expected, distance)
	}
}

func TestCalculateDistance_ZeroCoordinates(t *testing.T) {
	// Test with origin coordinates
	distance := CalculateDistance(0, 0, 0, 0)

	if distance != 0 {
		t.Errorf("Expected 0 distance for same origin, got: %f", distance)
	}
}

func TestCalculateDistance_NegativeCoordinates(t *testing.T) {
	// Test with negative coordinates (southern/western hemispheres)
	lat1 := -33.8688 // Sydney
	lon1 := 151.2093
	lat2 := -34.0522 // Another point in southern hemisphere
	lon2 := -118.2437

	distance := CalculateDistance(lat1, lon1, lat2, lon2)

	if distance < 0 {
		t.Errorf("Expected positive distance, got: %f", distance)
	}

	// Distance should be reasonable (not zero, not huge)
	if distance < 100 || distance > 20000 {
		t.Errorf("Unexpected distance: %f km", distance)
	}
}

func TestCalculateServerDistances_EmptyList(t *testing.T) {
	servers := []*types.Server{}
	userLat := 40.7128
	userLon := -74.0060

	// Should not panic
	CalculateServerDistances(servers, userLat, userLon)

	if len(servers) != 0 {
		t.Errorf("Expected empty list, got: %d servers", len(servers))
	}
}

func TestCalculateServerDistances_SingleServer(t *testing.T) {
	servers := []*types.Server{
		{
			URL:      "http://example.com/upload.php",
			Lat:      "34.0522",
			Lon:      "-118.2437",
			Name:     "Test Server",
			Country:  "USA",
			Sponsor:  "Test",
			ID:       "1",
			Host:     "example.com",
			Distance: 0,
		},
	}
	userLat := 40.7128
	userLon := -74.0060

	CalculateServerDistances(servers, userLat, userLon)

	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got: %d", len(servers))
	}

	// Distance should be calculated (not zero, since different locations)
	if servers[0].Distance == 0 {
		t.Errorf("Expected non-zero distance for different locations")
	}

	// Verify distance is reasonable
	if servers[0].Distance < 100 || servers[0].Distance > 10000 {
		t.Errorf("Unexpected distance: %f km", servers[0].Distance)
	}
}

func TestCalculateServerDistances_MultipleServers(t *testing.T) {
	servers := []*types.Server{
		{Lat: "34.0522", Lon: "-118.2437", Distance: 0}, // Los Angeles
		{Lat: "51.5074", Lon: "-0.1278", Distance: 0},   // London
		{Lat: "40.7128", Lon: "-74.0060", Distance: 0},  // New York (same as user)
	}
	userLat := 40.7128
	userLon := -74.0060

	CalculateServerDistances(servers, userLat, userLon)

	// Third server (same location) should have distance 0
	if servers[2].Distance != 0 {
		t.Errorf("Expected distance 0 for same location, got: %f", servers[2].Distance)
	}

	// Other servers should have non-zero distances
	for i, s := range servers {
		if i == 2 {
			continue // Skip the same-location server
		}
		if s.Distance <= 0 {
			t.Errorf("Server %d: Expected positive distance, got: %f", i, s.Distance)
		}
	}
}

func TestCalculateServerDistances_InvalidCoordinate(t *testing.T) {
	servers := []*types.Server{
		{
			URL:      "http://example.com/upload.php",
			Lat:      "invalid",
			Lon:      "-118.2437",
			Name:     "Test Server",
			Country:  "USA",
			Sponsor:  "Test",
			ID:       "1",
			Host:     "example.com",
			Distance: 0,
		},
	}
	userLat := 40.7128
	userLon := -74.0060

	// Should not panic with invalid coordinate
	CalculateServerDistances(servers, userLat, userLon)

	// Distance should remain 0 on error
	if servers[0].Distance != 0 {
		t.Errorf("Expected distance 0 for invalid coordinate, got: %f", servers[0].Distance)
	}
}

func TestSortServersByDistance_EmptyList(t *testing.T) {
	servers := []*types.Server{}

	// Should not panic
	SortServersByDistance(servers)

	if len(servers) != 0 {
		t.Errorf("Expected empty list, got: %d servers", len(servers))
	}
}

func TestSortServersByDistance_SingleServer(t *testing.T) {
	servers := []*types.Server{
		{Distance: 100, Name: "Single"},
	}

	SortServersByDistance(servers)

	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got: %d", len(servers))
	}
}

func TestSortServersByDistance_AlreadySorted(t *testing.T) {
	servers := []*types.Server{
		{Distance: 10, Name: "Close"},
		{Distance: 20, Name: "Medium"},
		{Distance: 30, Name: "Far"},
	}

	SortServersByDistance(servers)

	// Order should be preserved
	if servers[0].Distance != 10 || servers[1].Distance != 20 || servers[2].Distance != 30 {
		t.Errorf("Unexpected order: %v", servers)
	}
}

func TestSortServersByDistance_ReverseOrder(t *testing.T) {
	servers := []*types.Server{
		{Distance: 30, Name: "Far"},
		{Distance: 20, Name: "Medium"},
		{Distance: 10, Name: "Close"},
	}

	SortServersByDistance(servers)

	// Should be sorted ascending
	if servers[0].Distance != 10 || servers[1].Distance != 20 || servers[2].Distance != 30 {
		t.Errorf("Expected sorted ascending, got: %v", servers)
	}
}

func TestSortServersByDistance_RandomOrder(t *testing.T) {
	servers := []*types.Server{
		{Distance: 50, Name: "A"},
		{Distance: 10, Name: "B"},
		{Distance: 30, Name: "C"},
		{Distance: 20, Name: "D"},
		{Distance: 40, Name: "E"},
	}

	SortServersByDistance(servers)

	// Verify sorted order
	expectedOrder := []float64{10, 20, 30, 40, 50}
	for i, expected := range expectedOrder {
		if servers[i].Distance != expected {
			t.Errorf("Position %d: expected distance %f, got %f", i, expected, servers[i].Distance)
		}
	}
}

func TestSortServersByDistance_SameDistance(t *testing.T) {
	servers := []*types.Server{
		{Distance: 100, Name: "A"},
		{Distance: 100, Name: "B"},
		{Distance: 100, Name: "C"},
	}

	SortServersByDistance(servers)

	// All have same distance, order should be preserved (stable sort)
	// Our bubble sort is stable for equal elements
	if servers[0].Name != "A" || servers[1].Name != "B" || servers[2].Name != "C" {
		t.Errorf("Expected stable sort for equal distances, got: %v", servers)
	}
}
