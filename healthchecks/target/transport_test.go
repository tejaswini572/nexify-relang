package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPDeliveryEngineRetriesTransientStatus(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	engine := NewHTTPDeliveryEngine(time.Second, 3, 3)
	err := engine.Do(context.Background(), HTTPRequest{URL: server.URL, Headers: make(http.Header)})
	if err != nil {
		t.Fatalf("delivery returned %v, want retry then success", err)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("request count = %d, want 2", got)
	}
}

func TestHTTPDeliveryEngineDoesNotRetry404(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	err := NewHTTPDeliveryEngine(time.Second, 3, 3).Do(context.Background(), HTTPRequest{URL: server.URL, Headers: make(http.Header)})
	var deliveryErr *DeliveryError
	if !errors.As(err, &deliveryErr) || !deliveryErr.Permanent {
		t.Fatalf("error = %v, want permanent DeliveryError", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("404 request count = %d, want 1", got)
	}
}

func TestDiscordTransportUsesSlackEndpointAndPayload(t *testing.T) {
	var requestPath string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		if got := r.Header.Get("User-Agent"); got != "healthchecks.io" {
			t.Errorf("User-Agent = %q, want healthchecks.io", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	transport := DiscordTransport{WebhookURL: server.URL, SiteName: "Mychecks", IconURL: "https://example.test/logo.png", HTTP: NewHTTPDeliveryEngine(time.Second, 3, 1)}
	err := transport.Deliver(context.Background(), DeliveryEvent{
		CheckName: "nightly", ProjectName: "Alices Project", Status: "down", Reason: "success signal did not arrive on time, grace time passed",
		Schedule: "* * * * *", TimeZone: "UTC", PingCount: 42, LastPingText: "Success, 10 minutes ago", CloakedURL: "https://example.test/cloaked/key/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if requestPath != "/slack" {
		t.Fatalf("path = %q, want /slack", requestPath)
	}
	attachments, ok := payload["attachments"].([]any)
	if !ok || len(attachments) != 1 {
		t.Fatalf("attachments = %#v", payload["attachments"])
	}
	attachment := attachments[0].(map[string]any)
	if got := attachment["fallback"]; got != `The check "nightly" is DOWN.` {
		t.Fatalf("fallback = %#v", got)
	}
	fields := attachment["fields"].([]any)
	foundSchedule := false
	for _, raw := range fields {
		field := raw.(map[string]any)
		if field["title"] == "Schedule" {
			foundSchedule = field["value"] == `\* \* \* \* \*`
		}
	}
	if !foundSchedule {
		t.Fatalf("Discord schedule field was not escaped: %#v", fields)
	}
}

func TestTestDeliveryUsesDetachedChannelSnapshot(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/slack" {
			t.Errorf("path = %s, want /slack", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	a := newApp()
	a.reset()
	id := a.addChannel(channel{Kind: "discord", Value: `{"webhook":{"url":"` + server.URL + `"}}`})
	if err := a.deliverTestNotification(context.Background(), id, "http://target.test"); err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("delivery count = %d, want 1", got)
	}
}

func TestGotifyTransportRetries404AndBuildsReferenceRequest(t *testing.T) {
	var calls atomic.Int32
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nested/message" || r.URL.Query().Get("token") != "a token" {
			t.Errorf("Gotify request = %s, want /nested/message?token=a+token", r.URL.String())
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	priority := 9
	transport := GotifyTransport{
		Config: gotifyConfig{URL: server.URL + "/nested", Token: "a token", Priority: &priority, PriorityUp: &priority},
		HTTP:   NewHTTPDeliveryEngine(time.Second, 3, 3),
	}
	err := transport.Deliver(context.Background(), DeliveryEvent{CheckName: "Foo", Status: "down", Reason: "success signal did not arrive on time, grace time passed", CloakedURL: "https://example.test/cloaked/key/"})
	if err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("Gotify 404 retry count = %d, want 3", got)
	}
	if payload["title"] != "Foo is DOWN" || payload["priority"] != float64(9) {
		t.Fatalf("Gotify payload fields = %#v", payload)
	}
	extras := payload["extras"].(map[string]any)
	display := extras["client::display"].(map[string]any)
	if display["contentType"] != "text/markdown" {
		t.Fatalf("Gotify extras = %#v", extras)
	}
	if message := payload["message"].(string); message == "" || message[0:4] != "🔴" {
		t.Fatalf("Gotify message = %q", message)
	}
}

func TestGotifyCreationAndEditPersistStructuredConfiguration(t *testing.T) {
	a := newApp()
	a.reset()
	a.mu.Lock()
	projectID := a.project
	a.sessions["alice-session"] = "alice"
	a.mu.Unlock()
	post := func(path string, values url.Values) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "sessionid", Value: "alice-session"})
		req.AddCookie(&http.Cookie{Name: "csrftoken", Value: "csrf"})
		recorder := httptest.NewRecorder()
		a.serveHTTP(recorder, req)
		return recorder
	}
	form := url.Values{"url": {"example.test/gotify"}, "token": {"first-token"}, "priority": {"2"}, "priority_up": {"9"}, "csrfmiddlewaretoken": {"csrf"}}
	if got := post("/projects/"+projectID+"/add_gotify/", form).Code; got != http.StatusFound {
		t.Fatalf("Gotify create = %d, want 302", got)
	}
	a.mu.Lock()
	if len(a.channels) != 1 {
		a.mu.Unlock()
		t.Fatalf("channel count = %d, want 1", len(a.channels))
	}
	channelID := a.channels[0].ID
	a.mu.Unlock()
	created, ok := a.channelSnapshot(channelID)
	if !ok || created.ProjectID != projectID {
		t.Fatalf("created channel = %#v", created)
	}
	config, err := gotifyConfiguration(created.Value)
	if err != nil || config.URL != "https://example.test/gotify" || *config.Priority != 2 || *config.PriorityUp != 9 {
		t.Fatalf("created Gotify config = %#v, %v", config, err)
	}
	form = url.Values{"url": {"https://example.test/new"}, "token": {"second-token"}, "priority": {"5"}, "priority_up": {"0"}, "csrfmiddlewaretoken": {"csrf"}}
	if got := post("/integrations/"+channelID+"/edit/", form).Code; got != http.StatusFound {
		t.Fatalf("Gotify edit = %d, want 302", got)
	}
	edited, _ := a.channelSnapshot(channelID)
	config, err = gotifyConfiguration(edited.Value)
	if err != nil || config.Token != "second-token" || *config.Priority != 5 || *config.PriorityUp != 0 {
		t.Fatalf("edited Gotify config = %#v, %v", config, err)
	}
}
