package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestConcurrentPingPauseResume(t *testing.T) {
	a := newApp()
	a.reset()
	id := uuid()
	a.mu.Lock()
	a.checks[id] = &check{ID: id, Name: "concurrent", Slug: "concurrent", Timeout: 60, Grace: 60, Status: "new"}
	a.mu.Unlock()

	server := httptest.NewServer(http.HandlerFunc(a.serveHTTP))
	defer server.Close()

	request := func(method, path string, api bool) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if api {
			req.Header.Set("X-Api-Key", apiKey)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		resp.Body.Close()
	}

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(3)
		go func() { defer wg.Done(); request(http.MethodPost, "/ping/"+id, false) }()
		go func() { defer wg.Done(); request(http.MethodPost, "/api/v1/checks/"+id+"/pause", true) }()
		go func() { defer wg.Done(); request(http.MethodPost, "/api/v1/checks/"+id+"/resume", true) }()
	}
	wg.Wait()
}

func TestUniqueLookupRequiresEveryRequestedField(t *testing.T) {
	a := newApp()
	a.reset()
	a.mu.Lock()
	a.checks["one"] = &check{ID: "one", ProjectID: a.project, Name: "same", Slug: "same", Timeout: 60, Grace: 60}
	a.mu.Unlock()
	server := httptest.NewServer(http.HandlerFunc(a.serveHTTP))
	defer server.Close()

	post := func(body string) int {
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/checks/", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Api-Key", apiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}
	// Name matches an existing check but timeout does not, so this must create
	// rather than returning the existing check.
	if got := post(`{"name":"same","timeout":120,"grace":60,"unique":["name","timeout"]}`); got != http.StatusCreated {
		t.Fatalf("partial unique match status = %d, want 201", got)
	}
	if got := post(`{"name":"same","timeout":60,"grace":60,"unique":["name","timeout"]}`); got != http.StatusOK {
		t.Fatalf("complete unique match status = %d, want 200", got)
	}
}

func TestWriteAPIKeyMayBeSuppliedInJSONBody(t *testing.T) {
	a := newApp()
	a.reset()
	server := httptest.NewServer(http.HandlerFunc(a.serveHTTP))
	defer server.Close()
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/checks/", bytes.NewBufferString(`{"api_key":"`+apiKey+`","name":"body-key","timeout":60,"grace":60}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("body api key status = %d, want 201", resp.StatusCode)
	}
}

func TestBobMembershipAndCharlieProjectIsolation(t *testing.T) {
	a := newApp()
	a.reset()
	id := uuid()
	a.mu.Lock()
	a.checks[id] = &check{ID: id, ProjectID: a.project, Name: "alice-only", Slug: "alice-only", Timeout: 60, Grace: 60, Status: "new"}
	a.sessions["bob-session"] = "bob"
	a.sessions["charlie-session"] = "charlie"
	a.mu.Unlock()
	server := httptest.NewServer(http.HandlerFunc(a.serveHTTP))
	defer server.Close()

	// Membership applies to browser/project access: Bob can open Alice's check.
	req, err := http.NewRequest(http.MethodGet, server.URL+"/checks/"+id+"/details", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "bob-session"})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Bob member access = %d, want 200", resp.StatusCode)
	}

	// Charlie has a separate project and is forbidden from Alice's API object.
	req, err = http.NewRequest(http.MethodGet, server.URL+"/api/v1/checks/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Api-Key", charlieAPIKey)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Charlie API access = %d, want 403", resp.StatusCode)
	}

	if a.userCanAccess("charlie", *a.checks[id]) {
		t.Fatal("Charlie unexpectedly has Alice project membership")
	}
}

func TestNewChannelAssociatesWithExistingProjectChecks(t *testing.T) {
	a := newApp()
	a.reset()
	a.mu.Lock()
	a.checks["check"] = &check{ID: "check", ProjectID: a.project, Name: "associated", Timeout: 60, Grace: 60}
	a.mu.Unlock()
	channelID := a.addChannel(channel{Kind: "email", Value: "alerts@example.org", Down: true, Up: true})
	snapshot, ok := a.checkSnapshot("check")
	if !ok || len(snapshot.ChannelIDs) != 1 || snapshot.ChannelIDs[0] != channelID {
		t.Fatalf("check associations = %#v, want [%s]", snapshot.ChannelIDs, channelID)
	}
}

func TestComputedStatusAndSchedule(t *testing.T) {
	now := time.Now().UTC()
	last := now.Add(-61 * time.Second)
	c := check{Status: "up", Timeout: 60, Grace: 60, LastPing: &last}
	if got := c.currentStatus(now, false); got != "grace" {
		t.Fatalf("status in grace = %q, want grace", got)
	}
	last = now.Add(-121 * time.Second)
	if got := c.currentStatus(now, false); got != "down" {
		t.Fatalf("expired status = %q, want down", got)
	}
	last = time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	c = check{Status: "up", Schedule: "5 12 * * *", TZ: "UTC", Grace: 60, LastPing: &last}
	next := c.graceStart(last, false)
	if next == nil || next.Hour() != 12 || next.Minute() != 5 {
		t.Fatalf("cron next ping = %v, want 12:05 UTC", next)
	}
}

func TestScheduleNextRunKnownCalendarBoundaries(t *testing.T) {
	utc := time.UTC
	monthBoundary := time.Date(2026, time.January, 31, 23, 59, 0, 0, utc)
	got := nextCron("@monthly", monthBoundary, utc)
	want := time.Date(2026, time.February, 1, 0, 0, 0, 0, utc)
	if got == nil || !got.Equal(want) {
		t.Fatalf("@monthly next run = %v, want %v", got, want)
	}

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	// 01:30 EST on the 2026 spring-forward date. The following local midnight
	// is in EDT, so it is 04:00 UTC rather than 05:00 UTC.
	beforeDST := time.Date(2026, time.March, 8, 6, 30, 0, 0, utc)
	got = nextCron("@daily", beforeDST, loc)
	want = time.Date(2026, time.March, 9, 4, 0, 0, 0, utc)
	if got == nil || !got.Equal(want) {
		t.Fatalf("@daily after DST transition = %v, want %v", got, want)
	}

	onCalendarAfter := time.Date(2026, time.January, 31, 13, 0, 0, 0, utc)
	got = nextOnCalendar("12:34", onCalendarAfter, utc)
	want = time.Date(2026, time.February, 1, 12, 34, 0, 0, utc)
	if got == nil || !got.Equal(want) {
		t.Fatalf("OnCalendar month rollover = %v, want %v", got, want)
	}
}
