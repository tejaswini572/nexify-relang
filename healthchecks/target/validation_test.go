package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStrictNumericValidationAndAPIKeyErrors(t *testing.T) {
	a := newApp()
	a.reset()
	server := httptest.NewServer(http.HandlerFunc(a.serveHTTP))
	defer server.Close()

	request := func(key, payload string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/checks/", bytes.NewBufferString(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if key != "" {
			req.Header.Set("X-Api-Key", key)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}

	resp := request(apiKey, `{"name":"fraction","timeout":60.5,"grace":60}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("fractional timeout status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	resp = request("Z"+apiKey[1:], `{"name":"wrong","timeout":60,"grace":60}`)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unknown 32-character key status = %d, want 401", resp.StatusCode)
	}
	resp.Body.Close()
}
