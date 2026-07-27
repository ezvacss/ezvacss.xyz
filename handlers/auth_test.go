package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateSessionID(t *testing.T) {
	s1 := GenerateSessionID()
	s2 := GenerateSessionID()

	if len(s1) == 0 || len(s2) == 0 {
		t.Errorf("GenerateSessionID returned empty string")
	}

	if s1 == s2 {
		t.Errorf("GenerateSessionID generated duplicate session IDs: %s", s1)
	}
}

func TestCleanIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.1:8080", "192.168.1.1"},
		{"10.0.0.1", "10.0.0.1"},
		{"[::1]:8080", "::1"},
		{"::1", "::1"},
		{" 127.0.0.1 ", "127.0.0.1"},
	}

	for _, tt := range tests {
		got := cleanIP(tt.input)
		if got != tt.expected {
			t.Errorf("cleanIP(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestHandleLoginMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	rr := httptest.NewRecorder()

	HandleLogin(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d for GET /api/login, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleUserUnauthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	rr := httptest.NewRecorder()

	HandleUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["authenticated"] != false {
		t.Errorf("Expected authenticated=false for unauthenticated request, got %v", resp["authenticated"])
	}
}

func TestGetUserIDFromRequest(t *testing.T) {
	// Unauthenticated request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if id := GetUserIDFromRequest(req); id != 0 {
		t.Errorf("Expected UserID 0 for request without cookie, got %d", id)
	}

	// Invalid session cookie
	reqWithCookie := httptest.NewRequest(http.MethodGet, "/", nil)
	reqWithCookie.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid_session_12345"})
	if id := GetUserIDFromRequest(reqWithCookie); id != 0 {
		t.Errorf("Expected UserID 0 for invalid session cookie, got %d", id)
	}
}
