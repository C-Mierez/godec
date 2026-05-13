package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

func TestHealthHandlers_Live(t *testing.T) {
	// Arrange: Set up the handler
	handler := &HealthHandlers{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	// Act: Call the handler
	err := handler.Live(c)

	// Assert: Verify no error occurred
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert: Verify HTTP status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Assert: Verify response structure and content
	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != HealthStatusOK {
		t.Errorf("Expected status '%s', got '%s'", HealthStatusOK, response.Status)
	}

	// Verify timestamp is not empty and is in valid RFC3339 format
	if response.Timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}
	if _, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
		t.Errorf("Expected valid RFC3339 timestamp, got '%s': %v", response.Timestamp, err)
	}
}

func TestHealthHandlers_Ready(t *testing.T) {
	// Arrange: Set up the handler
	handler := &HealthHandlers{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, HealthReadyEndpoint, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	// Act: Call the handler
	err := handler.Ready(c)

	// Assert: Verify no error occurred
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert: Verify HTTP status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Assert: Verify response structure
	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != HealthStatusOK {
		t.Errorf("Expected status '%s', got '%s'", HealthStatusOK, response.Status)
	}
}
