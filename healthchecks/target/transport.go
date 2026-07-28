package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Transport delivers a check-status event to one notification provider.  It is
// deliberately independent from HTTP so providers which use another protocol
// can implement the same contract later.
type Transport interface {
	Deliver(context.Context, DeliveryEvent) error
}

// DeliveryEvent is the provider-neutral snapshot used for a notification. It
// contains values, not pointers into app state, so a delivery never keeps the
// store lock while doing network I/O.
type DeliveryEvent struct {
	CheckName    string
	ProjectName  string
	Description  string
	Tags         []string
	Status       string
	Reason       string
	DownDuration time.Duration
	Timeout      time.Duration
	Schedule     string
	TimeZone     string
	PingCount    int
	LastPingText string
	LastPingBody string
	CloakedURL   string
}

// HTTPDeliveryEngine owns the policy shared by JSON webhook transports. It is
// immutable after construction and safe to share between concurrent requests.
type HTTPDeliveryEngine struct {
	client       *http.Client
	timeout      time.Duration
	maxRedirects int
	attempts     int
}

// HTTPRequest describes one provider request. SuccessStatuses is intentionally
// supplied by the provider: some services accept a narrower or wider range.
type HTTPRequest struct {
	Method          string
	URL             string
	Headers         http.Header
	Body            []byte
	SuccessStatuses map[int]bool
	// Nil uses the shared webhook default (404 is permanent). An explicitly
	// empty map is useful for providers whose reference behavior retries 404.
	PermanentStatuses map[int]bool
}

// DeliveryError records whether another attempt could plausibly succeed.
// A 404 is permanent for webhook endpoints, matching Healthchecks' HTTP
// transport policy; all other transport failures are retried.
type DeliveryError struct {
	StatusCode int
	Err        error
	Permanent  bool
}

func (e *DeliveryError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("received status code %d", e.StatusCode)
	}
	return e.Err.Error()
}

func (e *DeliveryError) Unwrap() error { return e.Err }

func NewHTTPDeliveryEngine(timeout time.Duration, maxRedirects, attempts int) *HTTPDeliveryEngine {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if maxRedirects < 0 {
		maxRedirects = 3
	}
	if attempts < 1 {
		attempts = 3
	}
	client := &http.Client{Timeout: timeout}
	client.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		// `via` contains the requests already followed. Allow exactly three
		// redirects, just as the Django curl helper does.
		if len(via) > maxRedirects {
			return http.ErrUseLastResponse
		}
		return nil
	}
	return &HTTPDeliveryEngine{client: client, timeout: timeout, maxRedirects: maxRedirects, attempts: attempts}
}

var defaultHTTPDeliveryEngine = NewHTTPDeliveryEngine(30*time.Second, 3, 3)

func defaultSuccessStatuses() map[int]bool {
	return map[int]bool{http.StatusOK: true, http.StatusCreated: true, http.StatusAccepted: true, http.StatusNoContent: true}
}

// Do performs a request using the common timeout, redirect, and retry policy.
// It consumes and closes each response body before deciding whether to retry.
func (e *HTTPDeliveryEngine) Do(ctx context.Context, spec HTTPRequest) error {
	if spec.Method == "" {
		spec.Method = http.MethodPost
	}
	if len(spec.SuccessStatuses) == 0 {
		spec.SuccessStatuses = defaultSuccessStatuses()
	}
	for attempt := 0; attempt < e.attempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, e.timeout)
		req, err := http.NewRequestWithContext(attemptCtx, spec.Method, spec.URL, bytes.NewReader(spec.Body))
		if err == nil {
			req.Header = spec.Headers.Clone()
			resp, doErr := e.client.Do(req)
			if doErr == nil {
				_, readErr := io.Copy(io.Discard, resp.Body)
				closeErr := resp.Body.Close()
				if spec.SuccessStatuses[resp.StatusCode] && readErr == nil && closeErr == nil {
					cancel()
					return nil
				}
				if readErr != nil {
					err = readErr
				} else if closeErr != nil {
					err = closeErr
				} else {
					permanent := resp.StatusCode == http.StatusNotFound
					if spec.PermanentStatuses != nil {
						permanent = spec.PermanentStatuses[resp.StatusCode]
					}
					err = &DeliveryError{StatusCode: resp.StatusCode, Permanent: permanent}
				}
			} else {
				err = doErr
			}
		}
		cancel()
		var deliveryErr *DeliveryError
		if errors.As(err, &deliveryErr) && deliveryErr.Permanent {
			return err
		}
		if attempt == e.attempts-1 {
			return err
		}
	}
	return errors.New("delivery attempts exhausted")
}

func (e *HTTPDeliveryEngine) PostJSON(ctx context.Context, url string, payload any, headers http.Header, success map[int]bool) error {
	return e.PostJSONWithPolicy(ctx, url, payload, headers, success, nil)
}

// PostJSONWithPolicy lets a provider override the shared permanent-status
// policy. Passing an explicitly empty map means every non-success response is
// retryable, which is the behavior of Django's Gotify and ntfy transports.
func (e *HTTPDeliveryEngine) PostJSONWithPolicy(ctx context.Context, url string, payload any, headers http.Header, success, permanent map[int]bool) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if headers == nil {
		headers = make(http.Header)
	} else {
		headers = headers.Clone()
	}
	headers.Set("Content-Type", "application/json")
	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", "healthchecks.io")
	}
	return e.Do(ctx, HTTPRequest{Method: http.MethodPost, URL: url, Headers: headers, Body: body, SuccessStatuses: success, PermanentStatuses: permanent})
}

// DiscordTransport uses Discord's Slack-compatible incoming-webhook endpoint.
type DiscordTransport struct {
	WebhookURL string
	SiteName   string
	IconURL    string
	HTTP       *HTTPDeliveryEngine
}

func (t DiscordTransport) Deliver(ctx context.Context, event DeliveryEvent) error {
	engine := t.HTTP
	if engine == nil {
		engine = defaultHTTPDeliveryEngine
	}
	url := strings.Replace(t.WebhookURL, "https://discordapp.com/", "https://discord.com/", 1) + "/slack"
	return engine.PostJSON(ctx, url, t.payload(event), nil, defaultSuccessStatuses())
}

func (t DiscordTransport) payload(event DeliveryEvent) map[string]any {
	name := event.CheckName
	if name == "" {
		name = "TEST"
	}
	fields := make([]map[string]any, 0, 9)
	add := func(title, value string, short bool) {
		field := map[string]any{"title": title, "value": value}
		if short {
			field["short"] = true
		}
		fields = append(fields, field)
	}
	text := any(nil)
	if event.Reason != "" {
		text = "Reason: " + event.Reason + "."
	} else if event.Status == "up" && event.DownDuration > 0 {
		text = "The downtime lasted " + formatSentenceDuration(event.DownDuration) + "."
	}
	if event.Description != "" {
		add("Description", event.Description, false)
	}
	if event.ProjectName != "" {
		add("Project", event.ProjectName, true)
	}
	if len(event.Tags) > 0 {
		tags := make([]string, 0, len(event.Tags))
		for _, tag := range event.Tags {
			tags = append(tags, "`"+tag+"`")
		}
		add("Tags", strings.Join(tags, " "), true)
	}
	if event.Schedule == "" && event.Timeout > 0 {
		add("Period", formatSentenceDuration(event.Timeout), true)
	}
	if event.Schedule != "" {
		add("Schedule", strings.ReplaceAll(event.Schedule, "*", `\*`), true)
		add("Time Zone", event.TimeZone, true)
	}
	add("Total Pings", strconvItoa(event.PingCount), true)
	if event.LastPingText == "" {
		add("Last Ping", "Never", true)
	} else {
		add("Last Ping", event.LastPingText, true)
	}
	if event.LastPingBody != "" && !strings.Contains(event.LastPingBody, "```") {
		add("Last Ping Body", "```\n"+event.LastPingBody+"\n```", false)
	}
	color := "danger"
	if event.Status == "up" {
		color = "good"
	}
	return map[string]any{
		"username": t.SiteName,
		"icon_url": t.IconURL,
		"attachments": []any{map[string]any{
			"color": color, "fallback": fmt.Sprintf("The check %q is %s.", name, strings.ToUpper(event.Status)),
			"mrkdwn_in": []string{"fields"}, "title": fmt.Sprintf("“%s” is %s.", name, strings.ToUpper(event.Status)),
			"title_link": event.CloakedURL, "text": text, "fields": fields,
		}},
	}
}

type gotifyConfig struct {
	URL        string `json:"url"`
	Token      string `json:"token"`
	Priority   *int   `json:"priority,omitempty"`
	PriorityUp *int   `json:"priority_up,omitempty"`
}

// GotifyTransport posts the provider's Markdown payload to its application
// endpoint. Unlike Discord, Django retries a Gotify 404, so it passes an
// explicit empty permanent-status map to the common HTTP engine.
type GotifyTransport struct {
	Config gotifyConfig
	HTTP   *HTTPDeliveryEngine
}

func (t GotifyTransport) Deliver(ctx context.Context, event DeliveryEvent) error {
	priority := t.Config.Priority
	if event.Status == "up" {
		priority = t.Config.PriorityUp
	}
	if priority != nil && *priority == 0 {
		return nil
	}
	engine := t.HTTP
	if engine == nil {
		engine = defaultHTTPDeliveryEngine
	}
	url := strings.TrimRight(t.Config.URL, "/") + "/message?token=" + urlQueryEscape(t.Config.Token)
	payload := map[string]any{
		"title":   fmt.Sprintf("%s is %s", event.CheckName, strings.ToUpper(event.Status)),
		"message": gotifyMessage(event),
		"extras": map[string]any{
			"client::display": map[string]string{"contentType": "text/markdown"},
		},
	}
	if priority != nil {
		payload["priority"] = *priority
	}
	return engine.PostJSONWithPolicy(ctx, url, payload, nil, defaultSuccessStatuses(), map[int]bool{})
}

func gotifyMessage(event DeliveryEvent) string {
	name := event.CheckName
	if name == "" {
		name = "TEST"
	}
	if event.Status == "down" {
		message := fmt.Sprintf("🔴 The check [%s](%s) is **DOWN**", name, event.CloakedURL)
		if event.Reason != "" {
			message += " (" + event.Reason + ")"
		}
		return message + "."
	}
	return fmt.Sprintf("🟢 The check [%s](%s) is now **UP**.", name, event.CloakedURL)
}

func strconvItoa(n int) string { return fmt.Sprintf("%d", n) }

func formatSentenceDuration(d time.Duration) string {
	if d%(24*time.Hour) == 0 {
		n := int(d / (24 * time.Hour))
		if n == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", n)
	}
	if d%time.Hour == 0 {
		n := int(d / time.Hour)
		if n == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", n)
	}
	n := int(d / time.Minute)
	if n == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", n)
}

// deliverTestNotification is the one app entry point which invokes a
// transport. It snapshots the channel and project before calling the network;
// the app mutex is never held across retries or an outbound request.
func (a *app) deliverTestNotification(ctx context.Context, id, siteRoot string) error {
	a.mu.Lock()
	var selected channel
	found := false
	for _, ch := range a.channels {
		if ch.ID == id {
			selected = ch
			found = true
			break
		}
	}
	project := a.projects[selected.ProjectID]
	a.mu.Unlock()
	if !found {
		return errors.New("channel not found")
	}
	event := DeliveryEvent{
		CheckName: "TEST", ProjectName: project.Name, Status: "down",
		Reason:  "success signal did not arrive on time, grace time passed",
		Timeout: 24 * time.Hour, PingCount: 42, LastPingText: "Success, 1 day ago",
		CloakedURL: strings.TrimSuffix(siteRoot, "/") + "/cloaked/test/",
	}
	switch selected.Kind {
	case "discord":
		webhookURL, err := discordWebhookURL(selected.Value)
		if err != nil {
			return err
		}
		return DiscordTransport{
			WebhookURL: webhookURL,
			SiteName:   "Mychecks",
			IconURL:    strings.TrimSuffix(siteRoot, "/") + "/static/img/logo.png",
		}.Deliver(ctx, event)
	case "gotify":
		config, err := gotifyConfiguration(selected.Value)
		if err != nil {
			return err
		}
		return GotifyTransport{Config: config}.Deliver(ctx, event)
	default:
		return nil // Other providers do not have delivery transports yet.
	}
}

func discordWebhookURL(value string) (string, error) {
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value, nil
	}
	var config struct {
		Webhook struct {
			URL string `json:"url"`
		} `json:"webhook"`
	}
	if err := json.Unmarshal([]byte(value), &config); err != nil || config.Webhook.URL == "" {
		return "", errors.New("invalid Discord webhook configuration")
	}
	return config.Webhook.URL, nil
}

func gotifyConfiguration(value string) (gotifyConfig, error) {
	var config gotifyConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return gotifyConfig{}, err
	}
	if !validNotificationURL(config.URL) || config.Token == "" || len(config.Token) > 50 {
		return gotifyConfig{}, errors.New("invalid Gotify configuration")
	}
	for _, priority := range []*int{config.Priority, config.PriorityUp} {
		if priority != nil && (*priority < 0 || *priority > 9) {
			return gotifyConfig{}, errors.New("invalid Gotify priority")
		}
	}
	return config, nil
}

func urlQueryEscape(value string) string {
	// The token is a query value, not a URL fragment. QueryEscape's `+` form
	// encoding is what urllib.parse.urlencode (the reference implementation)
	// produces as well.
	return url.QueryEscape(value)
}
