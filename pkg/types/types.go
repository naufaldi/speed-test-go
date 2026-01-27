package types

import (
	"encoding/xml"
	"time"
)

// SpeedTestResult represents the final result of a speed test
type SpeedTestResult struct {
	Timestamp time.Time      `json:"timestamp"`
	Ping      PingResult     `json:"ping"`
	Download  TransferResult `json:"download"`
	Upload    TransferResult `json:"upload"`
	Server    *ServerInfo    `json:"server,omitempty"`
	Interface *InterfaceInfo `json:"interface,omitempty"`
	ISP       string         `json:"isp,omitempty"`
}

// PingResult contains ping/latency measurements
type PingResult struct {
	Jitter  float64 `json:"jitter"`  // milliseconds
	Latency float64 `json:"latency"` // milliseconds
}

// TransferResult contains download/upload measurements
type TransferResult struct {
	Bandwidth int64 `json:"bandwidth"` // bytes per second
	Bytes     int64 `json:"bytes"`
	Elapsed   int64 `json:"elapsed"` // milliseconds
}

// ServerInfo contains information about the test server
type ServerInfo struct {
	ID       string  `json:"id"`
	Host     string  `json:"host"`
	Name     string  `json:"name"`
	Location string  `json:"location"`
	Country  string  `json:"country"`
	Sponsor  string  `json:"sponsor"`
	Distance float64 `json:"distance"` // in km
}

// InterfaceInfo contains network interface information
type InterfaceInfo struct {
	InternalIP string `json:"internalIp"`
	ExternalIP string `json:"externalIp"`
	Name       string `json:"name"`
	MACAddr    string `json:"macAddr"`
}

// Server represents a speed test server from the JSON API
type Server struct {
	URL      string  `json:"url"`
	Lat      string  `json:"lat"`
	Lon      string  `json:"lon"`
	Name     string  `json:"name"`
	Country  string  `json:"country"`
	CC       string  `json:"cc"`
	Sponsor  string  `json:"sponsor"`
	ID       string  `json:"id"`
	Host     string  `json:"host"`
	Distance float64 `json:"distance"`
}

// ServerList represents the XML response from server list API
type ServerList struct {
	Servers []*Server `xml:"servers>server"`
}

// SettingsConfig represents the server-config element from speedtest.net API
type SettingsConfig struct {
	XMLName   xml.Name `xml:"server-config"`
	IP        string   `xml:"ip,attr"`
	Latitude  float64  `xml:"lat,attr"`
	Longitude float64  `xml:"lon,attr"`
	ISP       string   `xml:"isp,attr"`
}

// ClientConfig represents the client element from speedtest.net API
type ClientConfig struct {
	XMLName   xml.Name `xml:"client"`
	IP        string   `xml:"ip,attr"`
	Latitude  float64  `xml:"lat,attr"`
	Longitude float64  `xml:"lon,attr"`
	ISP       string   `xml:"isp,attr"`
	Country   string   `xml:"country,attr"`
}

// Settings represents the full settings response from speedtest.net API
type Settings struct {
	XMLName      xml.Name       `xml:"settings"`
	Version      string         `xml:"version,attr"`
	Mode         string         `xml:"mode,attr"`
	ServerConfig SettingsConfig `xml:"server-config"`
	Client       ClientConfig   `xml:"client"`
}

// UserLocation represents the user's detected location
type UserLocation struct {
	IP        string  `json:"ip"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ISP       string  `json:"isp"`
}

// TransferProgress represents progress during a transfer test
type TransferProgress struct {
	Rate         float64 // bytes per second
	BytesTotal   int64
	BytesCurrent int64
	Elapsed      time.Duration
}

// OutputState represents the current state of the test
type OutputState string

const (
	StateIdle     OutputState = "idle"
	StatePing     OutputState = "ping"
	StateDownload OutputState = "download"
	StateUpload   OutputState = "upload"
	StateDone     OutputState = "done"
)
