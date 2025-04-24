package clockify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/pterm/pterm"
)

type ClockifyAPI struct {
	apiKey      string
	workspaceId string
}

func NewClockifyAPI(apiKey, workspaceId string) *ClockifyAPI {
	return &ClockifyAPI{apiKey: apiKey, workspaceId: workspaceId}
}

type ClockifyTimeEntryPayload struct {
	Start       string `json:"start"`
	End         string `json:"end"`
	Description string `json:"description"`
	ProjectID   string `json:"projectId"`
	TaskID      string `json:"taskId"`
}

// Returns the clockify id of the new time entry
func (c *ClockifyAPI) PostNewTimeEntry(timeEntry *store.TimeEntry) (string, error) {
	// Format times in ISO 8601 format
	startTime := timeEntry.Start.Format(time.RFC3339)
	endTime := timeEntry.End.Format(time.RFC3339)

	// Create the time entry payload
	payload := ClockifyTimeEntryPayload{
		Start:       startTime,
		End:         endTime,
		Description: fmt.Sprintf("%s - %s", timeEntry.Project, timeEntry.Task),
		ProjectID:   "65ba4da699f4432f69476fef",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal time entry: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://api.clockify.me/api/v1/workspaces/%s/time-entries", c.workspaceId),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		pterm.Println(string(body))
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response to get the ID
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return result.ID, nil
}

func (c *ClockifyAPI) DeleteTimeEntry(timeEntry *store.TimeEntry) error {
	// Create HTTP request
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://api.clockify.me/api/v1/workspaces/{workspaceId}/time-entries/%s", timeEntry.ID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("X-Api-Key", c.apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
