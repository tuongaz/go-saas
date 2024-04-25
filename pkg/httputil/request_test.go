package httputil

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

// TestParseRequestBody tests the ParseRequestBody function.
func TestParseRequestBody(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
	}

	// Valid body
	jsonStr := []byte(`{"name":"test"}`)
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	got, err := ParseRequestBody[testStruct](req)
	if err != nil {
		t.Errorf("ParseRequestBody returned an error: %v", err)
	}
	if got.Name != "test" {
		t.Errorf("ParseRequestBody = %v, want %v", got.Name, "test")
	}

	// Invalid body
	jsonStr = []byte(`{"name":}`)
	req = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	_, err = ParseRequestBody[testStruct](req)
	if err == nil {
		t.Errorf("ParseRequestBody should have returned an error")
	}

	// Empty body
	req = httptest.NewRequest("POST", "/", nil)
	_, err = ParseRequestBody[testStruct](req)
	if err != nil {
		t.Errorf("ParseRequestBody returned an error for empty body: %v", err)
	}
}
