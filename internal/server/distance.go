package server

import (
	"math"
	"strconv"

	"github.com/user/speed-test-go/pkg/types"
)

// CalculateDistance calculates the great-circle distance between two points
// using the Haversine formula. Returns distance in kilometers.
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // km

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// CalculateServerDistances calculates the distance from user to each server
func CalculateServerDistances(servers []*types.Server, userLat, userLon float64) {
	for _, server := range servers {
		serverLat, err := parseCoordinate(server.Lat)
		if err != nil {
			server.Distance = 0
			continue
		}
		serverLon, err := parseCoordinate(server.Lon)
		if err != nil {
			server.Distance = 0
			continue
		}

		server.Distance = CalculateDistance(userLat, userLon, serverLat, serverLon)
	}
}

// SortServersByDistance sorts servers by distance (closest first)
func SortServersByDistance(servers []*types.Server) {
	// Simple bubble sort for small lists
	n := len(servers)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if servers[j].Distance > servers[j+1].Distance {
				servers[j], servers[j+1] = servers[j+1], servers[j]
			}
		}
	}
}

func parseCoordinate(coord string) (float64, error) {
	return strconv.ParseFloat(coord, 64)
}
