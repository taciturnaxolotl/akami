// Package wakatime provides a Go client for interacting with the WakaTime API.
// WakaTime is a time tracking service for programmers that automatically tracks
// how much time is spent on coding projects.
package wakatime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Default API URL for the WakaTime API v1
const (
	DefaultAPIURL = "https://api.wakatime.com/api/v1"
)

// Error types returned by the client
var (
	// ErrMarshalingHeartbeat occurs when a heartbeat can't be marshaled to JSON
	ErrMarshalingHeartbeat = fmt.Errorf("failed to marshal heartbeat to JSON")
	// ErrCreatingRequest occurs when the HTTP request cannot be created
	ErrCreatingRequest = fmt.Errorf("failed to create HTTP request")
	// ErrSendingRequest occurs when the HTTP request fails to send
	ErrSendingRequest = fmt.Errorf("failed to send HTTP request")
	// ErrInvalidStatusCode occurs when the API returns a non-success status code
	ErrInvalidStatusCode = fmt.Errorf("received invalid status code from API")
	// ErrDecodingResponse occurs when the API response can't be decoded
	ErrDecodingResponse = fmt.Errorf("failed to decode API response")
	// ErrUnauthorized occurs when the API rejects the provided credentials
	ErrUnauthorized = fmt.Errorf("unauthorized: invalid API key or insufficient permissions")
)

// Client represents a WakaTime API client with authentication and connection settings.
type Client struct {
	// APIKey is the user's WakaTime API key used for authentication
	APIKey string
	// APIURL is the base URL for the WakaTime API
	APIURL string
	// HTTPClient is the HTTP client used to make requests to the WakaTime API
	HTTPClient *http.Client
}

// NewClient creates a new WakaTime API client with the provided API key
// and a default HTTP client with a 10-second timeout.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		APIURL:     DefaultAPIURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewClientWithOptions creates a new WakaTime API client with the provided API key,
// custom API URL and a default HTTP client with a 10-second timeout.
func NewClientWithOptions(apiKey string, apiURL string) *Client {
	return &Client{
		APIKey:     apiKey,
		APIURL:     apiURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Heartbeat represents a coding activity heartbeat sent to the WakaTime API.
// Heartbeats are the core data structure for tracking time spent coding.
type Heartbeat struct {
	// Entity is the file path or resource being worked on
	Entity string `json:"entity"`
	// Type specifies the entity type (usually "file")
	Type string `json:"type"`
	// Time is the timestamp of the heartbeat in UNIX epoch format
	Time float64 `json:"time"`
	// Project is the optional project name associated with the entity
	Project string `json:"project,omitempty"`
	// Language is the optional programming language of the entity
	Language string `json:"language,omitempty"`
	// IsWrite indicates if the file was being written to (vs. just viewed)
	IsWrite bool `json:"is_write,omitempty"`
	// EditorName is the optional name of the editor or IDE being used
	EditorName string `json:"editor_name,omitempty"`
}

// StatusBarResponse represents the response from the WakaTime Status Bar API endpoint.
// This contains summary information about a user's coding activity for a specific time period.
type StatusBarResponse struct {
	// Data contains coding duration information
	Data struct {
		// GrandTotal contains the aggregated coding time information
		GrandTotal struct {
			// Text is the human-readable representation of the total coding time
			// Example: "3 hrs 42 mins"
			Text string `json:"text"`
			// TotalSeconds is the total time spent coding in seconds
			// This can be used for precise calculations or custom formatting
			TotalSeconds int `json:"total_seconds"`
		} `json:"grand_total"`
	} `json:"data"`
}

// SendHeartbeat sends a coding activity heartbeat to the WakaTime API.
// It returns an error if the request fails or returns a non-success status code.
func (c *Client) SendHeartbeat(heartbeat Heartbeat) error {
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalingHeartbeat, err)
	}

	req, err := http.NewRequest("POST", c.APIURL+"/users/current/heartbeats", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCreatingRequest, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: status code %d", ErrInvalidStatusCode, resp.StatusCode)
	}

	return nil
}

// GetStatusBar retrieves a user's current day coding activity summary from the WakaTime API.
// It returns an error if the request fails or returns a non-success status code.
func (c *Client) GetStatusBar() (StatusBarResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/current/statusbar/today", c.APIURL), nil)
	if err != nil {
		return StatusBarResponse{}, fmt.Errorf("%w: %v", ErrCreatingRequest, err)
	}

	req.Header.Set("Authorization", "Basic "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return StatusBarResponse{}, fmt.Errorf("%w: %v", ErrSendingRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return StatusBarResponse{}, ErrUnauthorized
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return StatusBarResponse{}, fmt.Errorf("%w: status code %d", ErrInvalidStatusCode, resp.StatusCode)
	}

	var durationResp StatusBarResponse
	if err := json.NewDecoder(resp.Body).Decode(&durationResp); err != nil {
		return StatusBarResponse{}, fmt.Errorf("%w: %v", ErrDecodingResponse, err)
	}

	return durationResp, nil
}
