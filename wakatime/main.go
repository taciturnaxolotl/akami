// Package wakatime provides a Go client for interacting with the WakaTime API.
// WakaTime is a time tracking service for programmers that automatically tracks
// how much time is spent on coding projects.
package wakatime

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
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
	// Branch is the optional git branch name
	Branch string `json:"branch,omitempty"`
	// Category is the optional activity category
	Category string `json:"category,omitempty"`
	// LineCount is the optional number of lines in the file
	LineCount int `json:"lines,omitempty"`
	// LineNo is the current line number
	LineNo int `json:"lineno,omitempty"`
	// CursorPos is the current column of text the cursor is on
	CursorPos int `json:"cursorpos,omitempty"`
	// UserAgent is the optional user agent string
	UserAgent string `json:"user_agent,omitempty"`
	// EntityType is the optional entity type (usually redundant with Type)
	EntityType string `json:"entity_type,omitempty"`
	// Dependencies is an optional list of project dependencies
	Dependencies []string `json:"dependencies,omitempty"`
	// ProjectRootCount is the optional number of directories in the project root path
	ProjectRootCount int `json:"project_root_count,omitempty"`
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
	// Set the user agent in the heartbeat data
	if heartbeat.UserAgent == "" {
		heartbeat.UserAgent = "wakatime/unset (" + runtime.GOOS + "-" + runtime.GOARCH + ") akami-wakatime/1.0.0"
	}

	data, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalingHeartbeat, err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/current/statusbar/today", c.APIURL), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCreatingRequest, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.APIKey)))
	// Set the user agent in the request header as well
	req.Header.Set("User-Agent", "wakatime/unset ("+runtime.GOOS+"-"+runtime.GOARCH+") akami-wakatime/1.0.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingRequest, err)
	}
	defer resp.Body.Close()

	// Read and log the response
	var respBody bytes.Buffer
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	respContent := respBody.String()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%w: %s", ErrUnauthorized, respContent)
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: status code %d, response: %s", ErrInvalidStatusCode, resp.StatusCode, respContent)
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

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.APIKey)))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return StatusBarResponse{}, fmt.Errorf("%w: %v", ErrSendingRequest, err)
	}
	defer resp.Body.Close()

	// Read the response body for potential error messages
	var respBody bytes.Buffer
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return StatusBarResponse{}, fmt.Errorf("failed to read response body: %v", err)
	}

	respContent := respBody.String()

	if resp.StatusCode == http.StatusUnauthorized {
		return StatusBarResponse{}, fmt.Errorf("%w: %s", ErrUnauthorized, respContent)
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return StatusBarResponse{}, fmt.Errorf("%w: status code %d, response: %s", ErrInvalidStatusCode, resp.StatusCode, respContent)
	}

	var durationResp StatusBarResponse
	if err := json.Unmarshal(respBody.Bytes(), &durationResp); err != nil {
		return StatusBarResponse{}, fmt.Errorf("%w: %v, response: %s", ErrDecodingResponse, err, respContent)
	}

	return durationResp, nil
}
