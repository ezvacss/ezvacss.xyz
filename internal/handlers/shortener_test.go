package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleShortenMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
	rr := httptest.NewRecorder()

	HandleShorten(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d for GET /shorten, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleShortenInvalidPayload(t *testing.T) {
	tests := []struct {
		name       string
		payload    map[string]string
		expectCode int
	}{
		{"Empty Payload", map[string]string{"url": ""}, http.StatusBadRequest},
		{"Invalid URL Format", map[string]string{"url": "not a url"}, http.StatusBadRequest},
		{"Self Domain Link", map[string]string{"url": "https://ezvacss.xyz/dashboard"}, http.StatusBadRequest},
		{"Non Existent Domain", map[string]string{"url": "http://opasdjmaspodsnmadasnmd.com/"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			HandleShorten(rr, req)

			if rr.Code != tt.expectCode {
				t.Errorf("HandleShorten(%s) code = %d, want %d. Body: %s", tt.name, rr.Code, tt.expectCode, rr.Body.String())
			}
		})
	}
}

func TestHandleUnshortenMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/unshorten", nil)
	rr := httptest.NewRecorder()

	HandleUnshorten(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d for POST /api/unshorten, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleUnshortenMissingCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/unshorten", nil)
	rr := httptest.NewRecorder()

	HandleUnshorten(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for missing code param, got %d", http.StatusBadRequest, rr.Code)
	}
}
