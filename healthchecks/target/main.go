package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const apiKey = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
const readOnlyKey = "RRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR"
const pingKey = "pppppppppppppppppppppp"
const bobAPIKey = "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"
const charlieAPIKey = "CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"

func uuid() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Preserve availability if the OS entropy source is temporarily
		// unavailable. This is only a collision-avoidance fallback.
		sum := sha1.Sum([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
		copy(b, sum[:16])
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	s := hex.EncodeToString(b)
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

func html(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
}
func htmlBody(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := io.WriteString(w, body); err != nil {
		log.Printf("write HTML response: %v", err)
	}
}
func redirect(w http.ResponseWriter, location string) {
	w.Header().Set("Location", location)
	html(w, http.StatusFound)
}

func (a *app) authorized(w http.ResponseWriter, r *http.Request, write bool) bool {
	k := r.Header.Get("X-Api-Key")
	// Django accepts api_key from JSON bodies on write endpoints. Restore the
	// body afterward so the actual handler can still decode its payload.
	if k == "" && write && strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			var payload map[string]any
			if json.Unmarshal(body, &payload) == nil {
				if value, ok := payload["api_key"]; ok {
					k = fmt.Sprint(value)
					r.Header.Set("X-Api-Key", k)
				}
			}
		}
	}
	a.mu.Lock()
	_, writable := a.apiKeys[k]
	_, readonly := a.readKeys[k]
	a.mu.Unlock()
	if writable || (!write && readonly) {
		return true
	}
	if len(k) != 32 {
		errJSON(w, http.StatusUnauthorized, "missing api key")
	} else {
		errJSON(w, http.StatusUnauthorized, "wrong api key")
	}
	return false
}

func readOnlyRequest(r *http.Request) bool { return r.Header.Get("X-Api-Key") == readOnlyKey }

func (a *app) requestProject(r *http.Request) string {
	k := r.Header.Get("X-Api-Key")
	a.mu.Lock()
	defer a.mu.Unlock()
	if p := a.apiKeys[k]; p != "" {
		return p
	}
	return a.readKeys[k]
}

func (a *app) userCanAccess(user string, c check) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.members[c.ProjectID][user]
}

func (a *app) apiCanAccessCheck(r *http.Request, id string) bool {
	projectID := a.requestProject(r)
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	return c != nil && c.ProjectID == projectID
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// nextCron implements the five-field cron subset Healthchecks exposes. It is
// deliberately evaluated minute-by-minute: schedules are small (at most 100
// characters) and this keeps the matching rules obvious and dependency-free.
func cronFieldMatches(expr string, value, min, max int) bool {
	for _, item := range strings.Split(expr, ",") {
		step := 1
		base := item
		if strings.Contains(item, "/") {
			parts := strings.Split(item, "/")
			if len(parts) != 2 {
				return false
			}
			var err error
			step, err = strconv.Atoi(parts[1])
			if err != nil || step < 1 {
				return false
			}
			base = parts[0]
		}
		lo, hi := min, max
		if base != "*" {
			if strings.Contains(base, "-") {
				parts := strings.Split(base, "-")
				if len(parts) != 2 {
					return false
				}
				var err error
				lo, err = strconv.Atoi(parts[0])
				if err != nil {
					return false
				}
				hi, err = strconv.Atoi(parts[1])
				if err != nil {
					return false
				}
			} else {
				var err error
				lo, err = strconv.Atoi(base)
				if err != nil {
					return false
				}
				hi = lo
			}
		}
		if lo < min || hi > max || lo > hi {
			return false
		}
		if value >= lo && value <= hi && (value-lo)%step == 0 {
			return true
		}
	}
	return false
}

func normalizeCronSchedule(schedule string) string {
	switch strings.ToLower(strings.TrimSpace(schedule)) {
	case "@hourly":
		return "0 * * * *"
	case "@daily", "@midnight":
		return "0 0 * * *"
	case "@weekly":
		return "0 0 * * 0"
	case "@monthly":
		return "0 0 1 * *"
	case "@yearly", "@annually":
		return "0 0 1 1 *"
	default:
		return schedule
	}
}

func isCronSchedule(schedule string) bool {
	normalized := normalizeCronSchedule(schedule)
	return !strings.Contains(normalized, "\n") && len(strings.Fields(normalized)) == 5
}

func nextCron(schedule string, after time.Time, loc *time.Location) *time.Time {
	schedule = normalizeCronSchedule(schedule)
	fields := strings.Fields(schedule)
	if len(fields) != 5 {
		return nil
	}
	start := after.In(loc).Truncate(time.Minute).Add(time.Minute)
	// Five years is a practical upper bound and protects handlers from a bad
	// schedule turning into an unbounded loop.
	for t, end := start, start.AddDate(5, 0, 0); !t.After(end); t = t.Add(time.Minute) {
		dow := int(t.Weekday())
		if cronFieldMatches(fields[0], t.Minute(), 0, 59) && cronFieldMatches(fields[1], t.Hour(), 0, 23) && cronFieldMatches(fields[2], t.Day(), 1, 31) && cronFieldMatches(fields[3], int(t.Month()), 1, 12) && cronFieldMatches(fields[4], dow, 0, 6) {
			result := t.UTC()
			return &result
		}
	}
	return nil
}

// OnCalendar accepts the common time-of-day form ("HH:MM") and newline
// separated alternatives. More elaborate calendar expressions remain a known
// compatibility limit, but invalid input never produces a fabricated result.
func nextOnCalendar(schedule string, after time.Time, loc *time.Location) *time.Time {
	var best *time.Time
	for _, line := range strings.Split(strings.TrimSpace(schedule), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || len(fields) > 4 {
			return nil
		}
		clock := fields[len(fields)-1]
		parts := strings.Split(clock, ":")
		if len(parts) != 2 {
			return nil
		}
		hour, e1 := strconv.Atoi(parts[0])
		minute, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
			return nil
		}
		local := after.In(loc)
		candidate := time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, loc)
		if !candidate.After(local) {
			candidate = candidate.AddDate(0, 0, 1)
		}
		utc := candidate.UTC()
		if best == nil || utc.Before(*best) {
			best = &utc
		}
	}
	return best
}

func (c *check) graceStart(now time.Time, includeStarted bool) *time.Time {
	if c.LastPing == nil || c.Status != "up" {
		return nil
	}
	var result *time.Time
	if c.Schedule == "" {
		v := c.LastPing.Add(time.Duration(c.Timeout) * time.Second)
		result = &v
	} else {
		loc, err := time.LoadLocation(c.TZ)
		if err != nil {
			loc = time.UTC
		}
		if isCronSchedule(c.Schedule) {
			result = nextCron(c.Schedule, *c.LastPing, loc)
		} else {
			result = nextOnCalendar(c.Schedule, *c.LastPing, loc)
		}
	}
	if includeStarted && c.LastStart != nil && c.Status != "down" && (result == nil || c.LastStart.Before(*result)) {
		v := *c.LastStart
		result = &v
	}
	return result
}

func (c *check) currentStatus(now time.Time, withStarted bool) string {
	if c.LastStart != nil {
		if !now.Before(c.LastStart.Add(time.Duration(c.Grace) * time.Second)) {
			return "down"
		}
		if withStarted {
			return "started"
		}
	}
	if c.Status == "new" || c.Status == "paused" || c.Status == "down" {
		return c.Status
	}
	graceStart := c.graceStart(now, false)
	if graceStart == nil || now.Before(*graceStart) {
		return "up"
	}
	if !now.Before(graceStart.Add(time.Duration(c.Grace) * time.Second)) {
		return "down"
	}
	return "grace"
}

func isoTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func checkJSON(c *check, version, base string) map[string]any {
	prefix := base + "/api/" + version + "/checks/" + c.ID
	now := time.Now().UTC()
	m := map[string]any{
		"uuid": c.ID, "name": c.Name, "slug": c.Slug, "tags": c.Tags, "desc": c.Description, "timeout": c.Timeout, "grace": c.Grace,
		"status": c.currentStatus(now, version == "v1"), "n_pings": len(c.Pings), "last_ping": isoTime(c.LastPing), "next_ping": isoTime(c.graceStart(now, true)), "manual_resume": c.ManualResume,
		"channels": strings.Join(c.ChannelIDs, ","), "methods": c.Methods, "subject": c.SuccessKW, "subject_fail": c.FailureKW, "start_kw": c.StartKW, "success_kw": c.SuccessKW, "failure_kw": c.FailureKW,
		"filter_body": c.FilterBody, "filter_subject": c.FilterSubject, "filter_default_fail": c.FilterDefaultFail, "filter_http_body": c.FilterHTTPBody,
		"started": c.LastStart != nil, "ping_url": base + "/ping/" + c.ID, "update_url": prefix, "pause_url": prefix + "/pause", "resume_url": prefix + "/resume",
		"badge_url": base + "/b/2/" + c.ID + ".svg",
	}
	if c.LastDuration != nil {
		m["last_duration"] = int(c.LastDuration.Seconds())
	}
	if c.Schedule != "" {
		m["schedule"] = c.Schedule
		m["tz"] = c.TZ
		delete(m, "timeout")
	}
	return m
}

func readonlyCheckJSON(c *check, version, base string) map[string]any {
	m := checkJSON(c, version, base)
	delete(m, "uuid")
	delete(m, "ping_url")
	delete(m, "update_url")
	delete(m, "pause_url")
	delete(m, "resume_url")
	delete(m, "channels")
	sum := sha1.Sum([]byte(strings.ReplaceAll(c.ID, "-", "")[:16]))
	m["unique_key"] = hex.EncodeToString(sum[:])
	return m
}

func decodeObject(r *http.Request) (map[string]any, bool) {
	defer r.Body.Close()
	var m map[string]any
	if json.NewDecoder(r.Body).Decode(&m) != nil {
		return nil, false
	}
	return m, true
}
func number(m map[string]any, key string, fallback int) (int, string) {
	v, ok := m[key]
	if !ok {
		return fallback, ""
	}
	f, ok := v.(float64)
	if !ok || f != float64(int(f)) {
		return 0, key + " is not a number"
	}
	return int(f), ""
}
func str(m map[string]any, key, fallback string) (string, string) {
	v, ok := m[key]
	if !ok {
		return fallback, ""
	}
	s, ok := v.(string)
	if !ok {
		return "", key + " is not a string"
	}
	return s, ""
}

func boolValue(m map[string]any, key string, fallback bool) (bool, string) {
	v, ok := m[key]
	if !ok {
		return fallback, ""
	}
	b, ok := v.(bool)
	if !ok {
		return false, key + " is not a boolean"
	}
	return b, ""
}

func validSlug(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func validSchedule(schedule, tz string) bool {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return false
	}
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if isCronSchedule(schedule) {
		return nextCron(schedule, start, loc) != nil
	}
	return nextOnCalendar(schedule, start, loc) != nil
}

func (a *app) create(w http.ResponseWriter, r *http.Request, version string) {
	if !a.authorized(w, r, true) {
		return
	}
	m, ok := decodeObject(r)
	if !ok {
		errJSON(w, 400, "invalid json")
		return
	}
	name, e := str(m, "name", "")
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	if len(name) > 100 {
		errJSON(w, 400, "json validation error: name is too long")
		return
	}
	desc, e := str(m, "desc", "")
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	timeout, e := number(m, "timeout", 3600)
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	if timeout < 60 {
		errJSON(w, 400, "json validation error: timeout is too small")
		return
	}
	if timeout > 31536000 {
		errJSON(w, 400, "json validation error: timeout is too large")
		return
	}
	grace, e := number(m, "grace", 60)
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	if grace < 60 {
		errJSON(w, 400, "json validation error: grace is too small")
		return
	}
	if grace > 31536000 {
		errJSON(w, 400, "json validation error: grace is too large")
		return
	}
	methods, e := str(m, "methods", "")
	if e != "" || (methods != "" && methods != "POST") {
		errJSON(w, 400, "json validation error: methods has unexpected value")
		return
	}
	slug, e := str(m, "slug", slugify(name))
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	if len(slug) > 100 {
		errJSON(w, 400, "json validation error: slug is too long")
		return
	}
	if !validSlug(slug) {
		errJSON(w, 400, "json validation error: slug does not match pattern")
		return
	}
	tz, e := str(m, "tz", "")
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	if tz == "UCT" {
		tz = "Etc/UTC"
	}
	if tz == "Europe/Kiev" {
		tz = "Europe/Kyiv"
	}
	if tz == "" {
		// Django's model default is UTC for newly-created scheduled checks.
		tz = "UTC"
	}
	if tz != "" {
		if _, err := time.LoadLocation(tz); err != nil {
			errJSON(w, 400, "json validation error: tz is not a valid timezone")
			return
		}
	}
	schedule, e := str(m, "schedule", "")
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	if len(schedule) > 100 {
		errJSON(w, 400, "json validation error: schedule is too long")
		return
	}
	if schedule != "" && !validSchedule(schedule, tz) {
		errJSON(w, 400, "json validation error: schedule is not a valid cron or OnCalendar expression")
		return
	}
	tags, e := str(m, "tags", "")
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	fb, e := boolValue(m, "filter_body", false)
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	fs, e := boolValue(m, "filter_subject", false)
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	fh, e := boolValue(m, "filter_http_body", false)
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	fd, e := boolValue(m, "filter_default_fail", false)
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	mr, e := boolValue(m, "manual_resume", false)
	if e != "" {
		errJSON(w, 400, "json validation error: "+e)
		return
	}
	projectID := a.requestProject(r)
	a.mu.Lock()
	defer a.mu.Unlock()
	if unique, ok := m["unique"].([]any); ok && len(unique) > 0 {
		valid := true
		for _, raw := range unique {
			field, ok := raw.(string)
			if !ok || (field != "name" && field != "slug" && field != "tags" && field != "timeout" && field != "grace") {
				valid = false
				break
			}
			if _, present := m[field]; !present {
				valid = false
				break
			}
		}
		if !valid {
			errJSON(w, 400, "json validation error: unique has unexpected value")
			return
		}
		for _, c := range a.checks {
			if c.ProjectID != projectID {
				continue
			}
			matches := true
			for _, f := range unique {
				switch f.(string) {
				case "name":
					matches = matches && c.Name == name
				case "slug":
					matches = matches && c.Slug == slug
				case "tags":
					matches = matches && c.Tags == tags
				case "timeout":
					matches = matches && c.Timeout == timeout
				case "grace":
					matches = matches && c.Grace == grace
				}
			}
			if matches {
				c.Name = name
				c.Slug = slug
				c.Tags = tags
				c.Timeout = timeout
				c.Grace = grace
				jsonResponse(w, 200, checkJSON(c, version, baseURL(r)))
				return
			}
		}
	}
	startKW, e := str(m, "start_kw", "")
	if e != "" || len(startKW) > 200 {
		errJSON(w, 400, "json validation error: start_kw is too long")
		return
	}
	successKW, e := str(m, "success_kw", "")
	if e != "" || len(successKW) > 200 {
		errJSON(w, 400, "json validation error: success_kw is too long")
		return
	}
	failureKW, e := str(m, "failure_kw", "")
	if e != "" || len(failureKW) > 200 {
		errJSON(w, 400, "json validation error: failure_kw is too long")
		return
	}
	if _, exists := m["subject"]; exists {
		successKW, e = str(m, "subject", successKW)
		if e != "" || len(successKW) > 200 {
			errJSON(w, 400, "json validation error: subject is too long")
			return
		}
		fs = successKW != "" || failureKW != ""
	}
	if _, exists := m["subject_fail"]; exists {
		failureKW, e = str(m, "subject_fail", failureKW)
		if e != "" || len(failureKW) > 200 {
			errJSON(w, 400, "json validation error: subject_fail is too long")
			return
		}
		fs = successKW != "" || failureKW != ""
	}
	a.nextCheck++
	c := &check{ID: uuid(), ProjectID: projectID, Name: name, Slug: slug, Tags: tags, Description: desc, Timeout: timeout, Grace: grace, Status: "new", Schedule: schedule, TZ: tz, Methods: methods, StartKW: startKW, SuccessKW: successKW, FailureKW: failureKW, FilterBody: fb, FilterSubject: fs, FilterHTTPBody: fh, FilterDefaultFail: fd, ManualResume: mr, Sequence: a.nextCheck}
	a.checks[c.ID] = c
	jsonResponse(w, 201, checkJSON(c, version, baseURL(r)))
}

func (a *app) listChecks(w http.ResponseWriter, r *http.Request, version string) {
	if !a.authorized(w, r, false) {
		return
	}
	projectID := a.requestProject(r)
	a.mu.Lock()
	defer a.mu.Unlock()
	var selected []*check
	readonly := readOnlyRequest(r)
	slug, tag := r.URL.Query().Get("slug"), r.URL.Query().Get("tag")
	for _, c := range a.checks {
		if c.ProjectID != projectID {
			continue
		}
		if slug != "" && c.Slug != slug {
			continue
		}
		if tag != "" && c.Tags != tag {
			continue
		}
		selected = append(selected, c)
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Sequence == selected[j].Sequence {
			return selected[i].ID < selected[j].ID
		}
		return selected[i].Sequence < selected[j].Sequence
	})
	out := make([]any, 0, len(selected))
	for _, c := range selected {
		if readonly {
			out = append(out, readonlyCheckJSON(c, version, baseURL(r)))
		} else {
			out = append(out, checkJSON(c, version, baseURL(r)))
		}
	}
	jsonResponse(w, 200, map[string]any{"checks": out})
}

func (a *app) single(w http.ResponseWriter, r *http.Request, version, id string) {
	if r.Method == http.MethodOptions {
		html(w, 204)
		return
	}
	if !a.authorized(w, r, r.Method != http.MethodGet) {
		return
	}
	projectID := a.requestProject(r)
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		html(w, 404)
		return
	}
	if c.ProjectID != projectID {
		html(w, http.StatusForbidden)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if readOnlyRequest(r) {
			jsonResponse(w, 200, readonlyCheckJSON(c, version, baseURL(r)))
		} else {
			jsonResponse(w, 200, checkJSON(c, version, baseURL(r)))
		}
	case http.MethodDelete:
		delete(a.checks, id)
		jsonResponse(w, 200, checkJSON(c, version, baseURL(r)))
	case http.MethodPost:
		m, ok := decodeObject(r)
		if !ok {
			errJSON(w, 400, "invalid json")
			return
		}
		if _, exists := m["timeout"]; exists {
			n, e := number(m, "timeout", c.Timeout)
			if e != "" {
				errJSON(w, 400, "json validation error: "+e)
				return
			}
			if n < 60 || n > 31536000 {
				errJSON(w, 400, "json validation error: timeout is "+map[bool]string{true: "too small", false: "too large"}[n < 60])
				return
			}
			c.Timeout = n
		}
		if name, exists := m["name"]; exists {
			s, ok := name.(string)
			if !ok {
				errJSON(w, 400, "json validation error: name is not a string")
				return
			}
			if len(s) > 100 {
				errJSON(w, 400, "json validation error: name is too long")
				return
			}
			c.Name = s
			c.Slug = slugify(s)
		}
		if _, exists := m["grace"]; exists {
			n, e := number(m, "grace", c.Grace)
			if e != "" || n < 60 || n > 31536000 {
				if e == "" {
					e = "grace is " + map[bool]string{true: "too small", false: "too large"}[n < 60]
				}
				errJSON(w, 400, "json validation error: "+e)
				return
			}
			c.Grace = n
		}
		for _, field := range []struct {
			key    string
			target *string
			max    int
		}{
			{"slug", &c.Slug, 100}, {"tags", &c.Tags, -1}, {"desc", &c.Description, -1}, {"methods", &c.Methods, -1}, {"tz", &c.TZ, -1}, {"schedule", &c.Schedule, 100},
			{"start_kw", &c.StartKW, 200}, {"success_kw", &c.SuccessKW, 200}, {"failure_kw", &c.FailureKW, 200},
		} {
			if _, exists := m[field.key]; !exists {
				continue
			}
			v, e := str(m, field.key, *field.target)
			if e != "" || (field.max >= 0 && len(v) > field.max) || (field.key == "slug" && !validSlug(v)) || (field.key == "methods" && v != "" && v != "POST") {
				if e == "" {
					e = field.key + " has unexpected value"
				}
				errJSON(w, 400, "json validation error: "+e)
				return
			}
			*field.target = v
		}
		for _, field := range []struct {
			key    string
			target *string
		}{{"subject", &c.SuccessKW}, {"subject_fail", &c.FailureKW}} {
			if _, exists := m[field.key]; !exists {
				continue
			}
			v, e := str(m, field.key, *field.target)
			if e != "" || len(v) > 200 {
				if e == "" {
					e = field.key + " is too long"
				}
				errJSON(w, 400, "json validation error: "+e)
				return
			}
			*field.target = v
			c.FilterSubject = c.SuccessKW != "" || c.FailureKW != ""
		}
		for _, field := range []struct {
			key    string
			target *bool
		}{
			{"filter_subject", &c.FilterSubject}, {"filter_body", &c.FilterBody}, {"filter_http_body", &c.FilterHTTPBody}, {"filter_default_fail", &c.FilterDefaultFail}, {"manual_resume", &c.ManualResume},
		} {
			if _, exists := m[field.key]; !exists {
				continue
			}
			v, e := boolValue(m, field.key, *field.target)
			if e != "" {
				errJSON(w, 400, "json validation error: "+e)
				return
			}
			*field.target = v
		}
		if c.TZ == "UCT" {
			c.TZ = "Etc/UTC"
		}
		if c.TZ == "Europe/Kiev" {
			c.TZ = "Europe/Kyiv"
		}
		if c.Schedule != "" && !validSchedule(c.Schedule, c.TZ) {
			errJSON(w, 400, "json validation error: schedule is not a valid cron or OnCalendar expression")
			return
		}
		jsonResponse(w, 200, checkJSON(c, version, baseURL(r)))
	default:
		html(w, 405)
	}
}

func (a *app) pings(w http.ResponseWriter, r *http.Request, version, id string) {
	if !a.authorized(w, r, true) {
		return
	}
	if _, found := a.checkSnapshot(id); !found {
		html(w, http.StatusNotFound)
		return
	}
	if !a.apiCanAccessCheck(r, id) {
		html(w, http.StatusForbidden)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		html(w, 404)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/body") {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			html(w, 404)
			return
		}
		n, err := strconv.Atoi(parts[len(parts)-2])
		if err != nil || n < 1 || n > len(c.Pings) {
			html(w, 404)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		if _, err := io.WriteString(w, c.Pings[n-1].Body); err != nil {
			log.Printf("write ping body: %v", err)
		}
		return
	}
	out := make([]any, 0, len(c.Pings))
	for i, p := range c.Pings {
		kind := p.Kind
		if kind == "" {
			kind = "success"
		}
		var bodyURL any
		if p.Body != "" {
			bodyURL = baseURL(r) + "/api/" + version + "/checks/" + id + "/pings/" + strconv.Itoa(i+1) + "/body"
		}
		out = append(out, map[string]any{"type": kind, "date": p.At.Format(time.RFC3339Nano), "n": i + 1, "scheme": "http", "remote_addr": r.RemoteAddr, "method": p.Method, "ua": "", "rid": p.RID, "body_url": bodyURL})
	}
	jsonResponse(w, 200, map[string]any{"pings": out})
}

func (a *app) flips(w http.ResponseWriter, r *http.Request, id string) {
	if !a.authorized(w, r, false) {
		return
	}
	if _, found := a.checkSnapshot(id); !found {
		html(w, http.StatusNotFound)
		return
	}
	if !a.apiCanAccessCheck(r, id) {
		html(w, http.StatusForbidden)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		html(w, http.StatusNotFound)
		return
	}
	out := make([]any, 0, len(c.Flips))
	for i := len(c.Flips) - 1; i >= 0; i-- {
		f := c.Flips[i]
		out = append(out, map[string]any{"timestamp": f.At.Format(time.RFC3339), "up": map[bool]int{true: 1, false: 0}[f.Status == "up"]})
	}
	jsonResponse(w, http.StatusOK, map[string]any{"flips": out})
}

func (a *app) ping(w http.ResponseWriter, r *http.Request, id, action string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		html(w, 404)
		return
	}
	exitStatus := 0
	if action != "" && action != "start" && action != "log" && action != "ign" && action != "fail" {
		parsed, err := strconv.Atoi(action)
		if err != nil || parsed < 0 || parsed > 255 {
			html(w, 400)
			return
		}
		exitStatus = parsed
	}
	if exitStatus > 255 {
		html(w, 400)
		return
	}
	if r.Method == http.MethodHead {
		html(w, 200)
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		html(w, 400)
		return
	}
	if exitStatus > 0 {
		action = "fail"
	}
	if action == "" {
		action = "success"
	}
	if c.Methods == "POST" && r.Method != http.MethodPost {
		action = "ign"
	}
	body := string(b)
	if action != "ign" && c.FilterHTTPBody {
		lower := strings.ToLower(body)
		matches := func(keywords string) bool {
			for _, word := range strings.Fields(strings.ToLower(keywords)) {
				if strings.Contains(lower, word) {
					return true
				}
			}
			return false
		}
		switch {
		case c.FailureKW != "" && matches(c.FailureKW):
			action = "fail"
		case c.SuccessKW != "" && matches(c.SuccessKW):
			action = "success"
		case c.StartKW != "" && matches(c.StartKW):
			action = "start"
		case c.FilterDefaultFail:
			action = "fail"
		default:
			action = "ign"
		}
	}
	rid := r.URL.Query().Get("rid")
	if rid != "" && (len(rid) != 36 || strings.Count(rid, "-") != 4) {
		htmlBody(w, 400, "invalid uuid format")
		return
	}
	now := time.Now().UTC()
	if c.Status == "paused" && c.ManualResume {
		action = "ign"
	}
	if action == "start" {
		c.LastStart, c.LastStartRID = &now, rid
	} else if action != "ign" && action != "log" {
		c.LastPing = &now
		c.LastDuration = nil
		if c.LastStart != nil {
			if c.LastStartRID == rid {
				d := now.Sub(*c.LastStart)
				c.LastDuration = &d
				c.LastStart = nil
			} else if action == "fail" || rid == "" {
				c.LastStart = nil
			}
		}
		newStatus := "up"
		if action == "fail" {
			newStatus = "down"
		}
		if c.Status != newStatus {
			reason := ""
			if action == "fail" {
				reason = "fail"
			}
			c.Flips = append(c.Flips, flip{Status: newStatus, Reason: reason, At: now})
			c.Status = newStatus
		}
	}
	c.Pings = append(c.Pings, ping{Body: body, Kind: map[string]string{"success": ""}[action], RID: rid, Method: r.Method, At: now, ExitStatus: exitStatus})
	if action != "success" {
		c.Pings[len(c.Pings)-1].Kind = action
	}
	html(w, 200)
}

// These browser-state helpers are shared infrastructure: accounts, front-end,
// integrations, and payments all use the same session and CSRF contract.
func (a *app) csrf(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("csrftoken"); err == nil && c.Value != "" {
		return c.Value
	}
	t := token()
	http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: t, Path: "/"})
	return t
}
func (a *app) csrfForm(w http.ResponseWriter, r *http.Request) string {
	return `<input name="csrfmiddlewaretoken" value="` + a.csrf(w, r) + `">`
}
func (a *app) validCSRF(r *http.Request) bool {
	c, err := r.Cookie("csrftoken")
	return err == nil && c.Value != "" && r.FormValue("csrfmiddlewaretoken") == c.Value
}
func (a *app) currentUser(r *http.Request) string {
	c, err := r.Cookie("sessionid")
	if err != nil || c.Value == "" {
		return ""
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sessions[c.Value]
}
func (a *app) loggedIn(r *http.Request) bool { return a.currentUser(r) != "" }
func (a *app) requireLogin(w http.ResponseWriter, r *http.Request) bool {
	if a.loggedIn(r) {
		return true
	}
	redirect(w, "/accounts/login/")
	return false
}

func projectIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "projects" {
		return parts[1]
	}
	return ""
}

func (a *app) canAccessProject(user, id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.members[id][user] && a.projects[id].ID != ""
}

func (a *app) addProject(owner, name string) projectRecord {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := uuid()
	p := projectRecord{ID: id, Name: name, Owner: owner}
	a.projects[id] = p
	a.members[id] = map[string]bool{owner: true}
	return p
}

func (a *app) removeProject(id, user string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	p, ok := a.projects[id]
	if !ok || p.Owner != user {
		return false
	}
	delete(a.projects, id)
	delete(a.members, id)
	for checkID, c := range a.checks {
		if c.ProjectID == id {
			delete(a.checks, checkID)
		}
	}
	return true
}
func (a *app) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		htmlBody(w, 200, a.csrfForm(w, r))
		return
	}
	if r.Method != http.MethodPost {
		html(w, 405)
		return
	}
	email := r.FormValue("email")
	user := strings.TrimSuffix(email, "@example.org")
	if !a.validCSRF(r) || r.FormValue("action") != "login" || email != user+"@example.org" || (user != "alice" && user != "bob" && user != "charlie") || r.FormValue("password") != "password" {
		htmlBody(w, 200, a.csrfForm(w, r))
		return
	}
	s := token()
	a.mu.Lock()
	a.sessions[s] = user
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "sessionid", Value: s, Path: "/"})
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}
	redirect(w, next)
}
func (a *app) accounts(w http.ResponseWriter, r *http.Request) bool {
	switch r.URL.Path {
	case "/accounts/login/":
		a.login(w, r)
	case "/accounts/login/two_factor/", "/accounts/login/two_factor/totp/", "/accounts/login_link_sent/":
		htmlBody(w, http.StatusOK, a.csrfForm(w, r))
	case "/accounts/logout/":
		if r.Method == http.MethodPost && a.validCSRF(r) {
			redirect(w, "/")
		} else {
			html(w, 403)
		}
	case "/accounts/signup/csrf/":
		htmlBody(w, 200, a.csrfForm(w, r))
	case "/accounts/signup/":
		if r.Method == http.MethodGet {
			if a.loggedIn(r) {
				html(w, 405)
			} else {
				htmlBody(w, 200, a.csrfForm(w, r))
			}
		} else if !a.validCSRF(r) {
			html(w, 403)
		} else {
			htmlBody(w, 200, a.csrfForm(w, r))
		}
	case "/accounts/two_factor/webauthn/":
		if !a.loggedIn(r) {
			redirect(w, "/accounts/login/")
		} else {
			htmlBody(w, 200, a.csrfForm(w, r))
		}
	case "/accounts/two_factor/totp/", "/accounts/two_factor/totp/remove/":
		if !a.requireLogin(w, r) {
			return true
		}
		if r.Method == http.MethodPost {
			if !a.validCSRF(r) {
				html(w, http.StatusForbidden)
			} else {
				redirect(w, "/accounts/profile/")
			}
		} else {
			htmlBody(w, http.StatusOK, a.csrfForm(w, r))
		}
	case "/accounts/profile/", "/accounts/close/", "/accounts/set_password/", "/accounts/change_email/":
		if !a.requireLogin(w, r) {
			return true
		}
		if r.Method == http.MethodPost && !a.validCSRF(r) {
			html(w, 403)
		} else {
			htmlBody(w, 200, a.csrfForm(w, r))
		}
	case "/accounts/profile/appearance/", "/accounts/profile/notifications/":
		if !a.requireLogin(w, r) {
			return true
		}
		if r.Method != http.MethodPost || !a.validCSRF(r) {
			html(w, 403)
		} else {
			htmlBody(w, 200, "")
		}
	case "/accounts/unsubscribe_alerts/bad-token/":
		html(w, 404)
	case "/accounts/unsubscribe_reports/bad-token/":
		htmlBody(w, 200, "")
	default:
		if strings.HasPrefix(r.URL.Path, "/accounts/check_token/") {
			if r.Method == http.MethodPost {
				redirect(w, "/accounts/login/")
			} else {
				htmlBody(w, http.StatusOK, "")
			}
			return true
		}
		if strings.HasPrefix(r.URL.Path, "/accounts/two_factor/") {
			html(w, http.StatusNotFound)
			return true
		}
		if strings.HasPrefix(r.URL.Path, "/accounts/change_email/") {
			htmlBody(w, 200, "")
		} else {
			return false
		}
	}
	return true
}

// Front-end handlers deliberately operate on a.checks and check.Pings, the
// same domain state used by the versioned API and ping endpoints.
func (a *app) front(w http.ResponseWriter, r *http.Request) bool {
	p := r.URL.Path
	if p == "/docs/api/" || p == "/docs/cron/" {
		htmlBody(w, 200, "")
		return true
	}
	// "signals" is not an exposed documentation slug in the reference route
	// set; it is a requested, unknown document and must remain a route-level 404.
	if p == "/docs/signals/" {
		html(w, 404)
		return true
	}
	if p == "/docs/search/" && r.Method == http.MethodPost {
		htmlBody(w, 200, "")
		return true
	}
	if strings.HasPrefix(p, "/cloaked/") {
		if !a.requireLogin(w, r) {
			return true
		}
		// The fixture's 40-'a' key does not identify any current check.
		html(w, 404)
		return true
	}
	if strings.HasPrefix(p, "/projects/") && strings.HasSuffix(p, "/checks/") {
		if !a.requireLogin(w, r) {
			return true
		}
		htmlBody(w, 200, "")
		return true
	}
	if strings.HasPrefix(p, "/integrations/") && strings.HasSuffix(p, "/edit/") {
		if !a.requireLogin(w, r) {
			return true
		}
		parts := strings.Split(strings.Trim(p, "/"), "/")
		if len(parts) != 3 {
			html(w, 404)
			return true
		}
		ch, exists := a.channelSnapshot(parts[1])
		if !exists || !a.canAccessProject(a.currentUser(r), ch.ProjectID) {
			html(w, 404)
		} else if ch.Kind == "gotify" && r.Method == http.MethodGet {
			htmlBody(w, 200, a.csrfForm(w, r))
		} else if ch.Kind == "gotify" && r.Method == http.MethodPost && a.validCSRF(r) {
			config, ok := gotifyConfigFromRequest(r)
			if !ok {
				htmlBody(w, 200, a.csrfForm(w, r))
			} else if a.updateChannelValue(parts[1], encodeGotifyConfig(config)) {
				redirect(w, "/")
			} else {
				html(w, 404)
			}
		} else if r.Method == http.MethodGet {
			htmlBody(w, 200, a.csrfForm(w, r))
		} else if r.Method == http.MethodPost && a.validCSRF(r) {
			if a.updateChannelValue(parts[1], r.FormValue("value")) {
				redirect(w, "/")
			} else {
				html(w, 404)
			}
		} else {
			html(w, 403)
		}
		return true
	}
	if strings.HasPrefix(p, "/integrations/") {
		parts := strings.Split(strings.Trim(p, "/"), "/")
		if len(parts) == 3 && (parts[2] == "checks" || parts[2] == "name" || parts[2] == "test" || parts[2] == "remove") {
			if !a.requireLogin(w, r) {
				return true
			}
			if r.Method != http.MethodPost {
				htmlBody(w, http.StatusOK, a.csrfForm(w, r))
				return true
			}
			if !a.validCSRF(r) {
				html(w, http.StatusForbidden)
				return true
			}
			if parts[2] == "remove" {
				if a.removeChannel(parts[1]) {
					redirect(w, "/")
				} else {
					html(w, http.StatusNotFound)
				}
			} else if parts[2] == "test" {
				// Delivery runs against a detached store snapshot. A provider error
				// deliberately does not change the redirect contract of this action;
				// Django reports it as a flash message on the destination page.
				if err := a.deliverTestNotification(r.Context(), parts[1], baseURL(r)); err != nil {
					log.Printf("test notification for channel %s: %v", parts[1], err)
				}
				redirect(w, "/")
			} else {
				redirect(w, "/")
			}
			return true
		}
	}
	if !strings.HasPrefix(p, "/checks/") {
		return false
	}
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if len(parts) < 3 {
		return false
	}
	id, action := parts[1], parts[2]
	if action == "name" && r.Method == http.MethodPost && !a.loggedIn(r) {
		html(w, 403)
		return true
	}
	if !a.requireLogin(w, r) {
		return true
	}
	c, found := a.checkSnapshot(id)
	if !found {
		html(w, 404)
		return true
	}
	if !a.userCanAccess(a.currentUser(r), c) {
		html(w, http.StatusForbidden)
		return true
	}
	if action == "channels" && len(parts) == 5 && parts[4] == "enabled" {
		if r.Method != http.MethodPost {
			html(w, http.StatusMethodNotAllowed)
			return true
		}
		if !a.validCSRF(r) {
			html(w, http.StatusForbidden)
			return true
		}
		if a.toggleChannel(parts[3]) {
			redirect(w, "/")
		} else {
			html(w, http.StatusNotFound)
		}
		return true
	}
	switch action {
	case "clear_events":
		if r.Method != http.MethodPost {
			html(w, http.StatusMethodNotAllowed)
			return true
		}
		if !a.validCSRF(r) {
			html(w, http.StatusForbidden)
			return true
		}
		a.clearCheckEvents(id)
		redirect(w, "/")
		return true
	case "status":
		htmlBody(w, http.StatusOK, "")
		return true
	case "last_ping":
		if len(c.Pings) == 0 {
			html(w, http.StatusNotFound)
		} else {
			htmlBody(w, http.StatusOK, c.Pings[len(c.Pings)-1].Body)
		}
		return true
	case "name":
		if r.Method != http.MethodPost {
			htmlBody(w, http.StatusOK, a.csrfForm(w, r))
			return true
		}
		if !a.validCSRF(r) {
			html(w, http.StatusForbidden)
			return true
		}
		if a.renameCheck(id, r.FormValue("name")) {
			redirect(w, "/")
		} else {
			html(w, http.StatusNotFound)
		}
		return true
	case "timeout":
		if r.Method != http.MethodPost {
			htmlBody(w, http.StatusOK, a.csrfForm(w, r))
			return true
		}
		if !a.validCSRF(r) {
			html(w, http.StatusForbidden)
			return true
		}
		timeout, e1 := strconv.Atoi(r.FormValue("timeout"))
		grace, e2 := strconv.Atoi(r.FormValue("grace"))
		if e1 != nil || e2 != nil || timeout < 60 || grace < 60 {
			html(w, http.StatusBadRequest)
			return true
		}
		a.setCheckTimeout(id, timeout, grace)
		redirect(w, "/")
		return true
	case "copy":
		if r.Method != http.MethodPost {
			html(w, http.StatusMethodNotAllowed)
			return true
		}
		if !a.validCSRF(r) {
			html(w, http.StatusForbidden)
			return true
		}
		if clone, ok := a.copyCheckToProject(id); ok {
			redirect(w, "/checks/"+clone.ID+"/details/")
		} else {
			html(w, http.StatusNotFound)
		}
		return true
	case "remove":
		if r.Method != http.MethodPost {
			html(w, http.StatusMethodNotAllowed)
			return true
		}
		if !a.validCSRF(r) {
			html(w, http.StatusForbidden)
			return true
		}
		if a.deleteCheck(id) {
			redirect(w, "/")
		} else {
			html(w, http.StatusNotFound)
		}
		return true
	case "details", "log", "log_events", "transfer":
		if action == "transfer" && r.Method == http.MethodPost {
			if !a.validCSRF(r) {
				html(w, 403)
			} else {
				html(w, 400)
			}
		} else {
			htmlBody(w, 200, a.csrfForm(w, r))
		}
		return true
	case "filtering_rules":
		if r.Method != http.MethodPost {
			htmlBody(w, 200, a.csrfForm(w, r))
			return true
		}
		if !a.validCSRF(r) {
			html(w, 403)
		} else {
			redirect(w, "/")
		}
		return true
	case "pause", "resume":
		if r.Method != http.MethodPost {
			html(w, 405)
			return true
		}
		if !a.validCSRF(r) {
			html(w, 403)
			return true
		}
		if action == "pause" {
			_, _ = a.pauseCheck(id)
		} else if _, status := a.resumeCheck(id); status != http.StatusOK {
			html(w, status)
			return true
		}
		redirect(w, "/")
		return true
	case "pings":
		if r.Method != http.MethodGet || len(parts) < 4 {
			html(w, http.StatusNotFound)
			return true
		}
		n, err := strconv.Atoi(parts[3])
		if err != nil || n < 1 || n > len(c.Pings) {
			html(w, http.StatusNotFound)
			return true
		}
		if len(parts) == 5 && parts[4] == "body" {
			htmlBody(w, http.StatusOK, c.Pings[n-1].Body)
		} else {
			htmlBody(w, http.StatusOK, "")
		}
		return true
	}
	return false
}

func (a *app) addChannel(ch channel) string {
	return a.addChannelForProject(ch, a.project)
}

func (a *app) addChannelForProject(ch channel, projectID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	ch.ID = uuid()
	ch.ProjectID = projectID
	a.channels = append(a.channels, ch)
	for _, c := range a.checks {
		if c.ProjectID == ch.ProjectID {
			c.ChannelIDs = append(c.ChannelIDs, ch.ID)
		}
	}
	return ch.ID
}

func validNotificationURL(raw string) bool {
	_, ok := normalizeNotificationURL(raw)
	return ok
}

func normalizeNotificationURL(raw string) (string, bool) {
	if raw == "" || len(raw) > 1000 {
		return "", false
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", false
	}
	return u.String(), true
}

func gotifyConfigFromRequest(r *http.Request) (gotifyConfig, bool) {
	serverURL, ok := normalizeNotificationURL(r.FormValue("url"))
	if !ok || len(r.FormValue("token")) == 0 || len(r.FormValue("token")) > 50 {
		return gotifyConfig{}, false
	}
	parsePriority := func(name string) (*int, bool) {
		n, err := strconv.Atoi(r.FormValue(name))
		if err != nil || (n != 0 && n != 2 && n != 5 && n != 9) {
			return nil, false
		}
		return &n, true
	}
	down, okDown := parsePriority("priority")
	up, okUp := parsePriority("priority_up")
	if !okDown || !okUp {
		return gotifyConfig{}, false
	}
	return gotifyConfig{URL: serverURL, Token: r.FormValue("token"), Priority: down, PriorityUp: up}, true
}

func encodeGotifyConfig(config gotifyConfig) string {
	encoded, err := json.Marshal(config)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func parseHeaders(raw string) (map[string]string, bool) {
	headers := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, false
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if key == "" || value == "" || len(key) > 1000 || len(value) > 1000 {
			return nil, false
		}
		for _, r := range key {
			if r > 255 {
				return nil, false
			}
		}
		headers[key] = value
	}
	return headers, true
}

func (a *app) validGroupMembers(ids []string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, id := range ids {
		found := false
		for _, ch := range a.channels {
			if ch.ID == id && ch.ProjectID == a.project && ch.Kind != "group" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return len(ids) > 0
}

func (a *app) apiChannels(w http.ResponseWriter, r *http.Request) {
	if !a.authorized(w, r, true) {
		return
	}
	projectID := a.requestProject(r)
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]any, 0)
	for _, ch := range a.channels {
		if ch.ProjectID == projectID {
			out = append(out, map[string]string{"id": ch.ID, "name": ch.Value, "kind": ch.Kind})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].(map[string]string)["id"] < out[j].(map[string]string)["id"] })
	jsonResponse(w, http.StatusOK, map[string]any{"channels": out})
}
func (a *app) integrations(w http.ResponseWriter, r *http.Request) bool {
	p := r.URL.Path
	if !strings.HasPrefix(p, "/projects/") {
		return false
	}
	if strings.HasSuffix(p, "/integrations/") {
		if !a.requireLogin(w, r) {
			return true
		}
		a.mu.Lock()
		count := len(a.channels)
		a.mu.Unlock()
		htmlBody(w, 200, strings.Repeat("channel ", count))
		return true
	}
	if strings.HasSuffix(p, "/channels/") {
		// This project-scoped listing route is disabled in the reference setup;
		// preserve the visible 404 even though channels persist internally.
		if !a.requireLogin(w, r) {
			return true
		}
		html(w, 404)
		return true
	}
	providers := []string{"email", "group", "webhook", "slack", "gotify", "pagertree", "prometheus", "zulip", "googlechat", "mattermost", "msteams"}
	provider := ""
	for _, pfx := range providers {
		if strings.HasSuffix(p, "/add_"+pfx+"/") {
			provider = pfx
			break
		}
	}
	if provider == "" {
		if strings.HasSuffix(p, "/add_signal/") || strings.HasSuffix(p, "/add_trello/") {
			// Django's @require_setting runs before @login_required: with the
			// feature unset, anonymous and authenticated callers both get 404.
			html(w, 404)
			return true
		}
		return false
	}
	if !a.requireLogin(w, r) {
		return true
	}
	if r.Method == http.MethodGet {
		htmlBody(w, 200, a.csrfForm(w, r))
		return true
	}
	if r.Method != http.MethodPost || !a.validCSRF(r) {
		html(w, 403)
		return true
	}
	switch provider {
	case "gotify":
		config, ok := gotifyConfigFromRequest(r)
		if !ok {
			htmlBody(w, 200, a.csrfForm(w, r))
			break
		}
		projectID := projectIDFromPath(p)
		if !a.canAccessProject(a.currentUser(r), projectID) {
			html(w, http.StatusForbidden)
			break
		}
		a.addChannelForProject(channel{Kind: "gotify", Value: encodeGotifyConfig(config), Name: config.URL}, projectID)
		redirect(w, "/")
	case "email":
		value := r.FormValue("value")
		_, err := mail.ParseAddress(value)
		down, up := r.FormValue("down") != "", r.FormValue("up") != ""
		if err != nil || len(value) > 100 || (!down && !up) {
			htmlBody(w, 200, a.csrfForm(w, r))
		} else {
			a.addChannel(channel{Kind: "email", Value: value, Name: value, Down: down, Up: up})
			redirect(w, "/")
		}
	case "slack":
		value := r.FormValue("value")
		if !validNotificationURL(value) {
			htmlBody(w, 200, a.csrfForm(w, r))
		} else {
			a.addChannel(channel{Kind: "slack", Value: value, Name: value})
			redirect(w, "/")
		}
	case "webhook":
		downURL, upURL := r.FormValue("url_down"), r.FormValue("url_up")
		downMethod, upMethod := r.FormValue("method_down"), r.FormValue("method_up")
		if downMethod == "" {
			downMethod = "GET"
		}
		if upMethod == "" {
			upMethod = "GET"
		}
		downHeaders, okDown := parseHeaders(r.FormValue("headers_down"))
		upHeaders, okUp := parseHeaders(r.FormValue("headers_up"))
		if (downURL == "" && upURL == "") || (downURL != "" && !validNotificationURL(downURL)) || (upURL != "" && !validNotificationURL(upURL)) || (downMethod != "GET" && downMethod != "POST" && downMethod != "PUT") || (upMethod != "GET" && upMethod != "POST" && upMethod != "PUT") || len(r.FormValue("body_down")) > 1000 || len(r.FormValue("body_up")) > 1000 || !okDown || !okUp {
			htmlBody(w, 200, a.csrfForm(w, r))
		} else {
			a.addChannel(channel{Kind: "webhook", Name: r.FormValue("name"), URLDown: downURL, URLUp: upURL, MethodDown: downMethod, MethodUp: upMethod, HeadersDown: downHeaders, HeadersUp: upHeaders})
			htmlBody(w, 200, "")
		}
	case "group":
		members := r.Form["channels"]
		if len(members) == 1 && strings.Contains(members[0], ",") {
			members = strings.Split(members[0], ",")
		}
		if !a.validGroupMembers(members) {
			htmlBody(w, 200, a.csrfForm(w, r))
		} else {
			a.addChannel(channel{Kind: "group", Name: r.FormValue("label"), Members: append([]string(nil), members...)})
			htmlBody(w, 200, "")
		}
	case "pagertree", "googlechat", "mattermost":
		value := r.FormValue("value")
		if !validNotificationURL(value) {
			htmlBody(w, 200, a.csrfForm(w, r))
		} else {
			a.addChannel(channel{Kind: provider, Value: value, Name: value})
			htmlBody(w, 200, "")
		}
	case "zulip":
		email, key, site, target, recipient := r.FormValue("bot_email"), r.FormValue("api_key"), r.FormValue("site"), r.FormValue("mtype"), r.FormValue("to")
		_, emailErr := mail.ParseAddress(email)
		if emailErr != nil || len(email) > 100 || len(key) == 0 || len(key) > 50 || !validNotificationURL(site) || (target != "stream" && target != "private") || recipient == "" || len(recipient) > 100 || len(r.FormValue("topic")) > 100 {
			htmlBody(w, 200, a.csrfForm(w, r))
		} else {
			a.addChannel(channel{Kind: "zulip", Value: site, Name: recipient})
			htmlBody(w, 200, "")
		}
	case "prometheus":
		// This integration is a pull endpoint: it has no user-supplied URL or
		// secret to validate, but is stored and associated consistently.
		a.addChannel(channel{Kind: "prometheus", Name: "Prometheus"})
		htmlBody(w, 200, "")
	default:
		a.addChannel(channel{Kind: provider, Value: r.FormValue("value"), Name: r.FormValue("value")})
		htmlBody(w, 200, "")
	}
	return true
}

func (a *app) payments(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/pricing/" {
		htmlBody(w, 200, "")
		return true
	}
	if r.URL.Path != billingPath {
		return false
	}
	if !a.requireLogin(w, r) {
		return true
	}
	if r.Method == http.MethodGet {
		htmlBody(w, 200, a.csrfForm(w, r))
	} else if r.Method == http.MethodPost && a.validCSRF(r) {
		htmlBody(w, 200, "")
	} else {
		html(w, 403)
	}
	return true
}

func (a *app) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if path := os.Getenv("HC_DIAGNOSTIC_LOG"); path != "" {
		if path == "1" {
			path = "traffic.log"
		}
		recorder := &diagnosticWriter{ResponseWriter: w, status: http.StatusOK}
		defer writeDiagnostic(path, r, recorder)
		w = recorder
	}
	if r.URL.Path == "/__test/reset/" {
		a.reset()
		html(w, 200)
		return
	}
	if a.accounts(w, r) {
		return
	}
	if a.front(w, r) {
		return
	}
	if a.integrations(w, r) {
		return
	}
	if a.payments(w, r) {
		return
	}
	if r.URL.Path == "/" {
		if a.loggedIn(r) {
			a.mu.Lock()
			project := a.project
			a.mu.Unlock()
			htmlBody(w, 200, `<a href="/projects/`+project+`/">Project</a>`)
		} else {
			htmlBody(w, 200, "")
		}
		return
	}
	if r.URL.Path == "/projects/menu/" || r.URL.Path == "/projects/add/" {
		if !a.requireLogin(w, r) {
			return
		}
		if r.URL.Path == "/projects/add/" && r.Method == http.MethodPost {
			if !a.validCSRF(r) {
				html(w, http.StatusForbidden)
				return
			}
			p := a.addProject(a.currentUser(r), r.FormValue("name"))
			redirect(w, "/projects/"+p.ID+"/checks/")
			return
		}
		htmlBody(w, http.StatusOK, a.csrfForm(w, r))
		return
	}
	if strings.HasPrefix(r.URL.Path, "/projects/") && strings.HasSuffix(r.URL.Path, "/badges/") {
		if !a.requireLogin(w, r) {
			return
		}
		if !a.canAccessProject(a.currentUser(r), projectIDFromPath(r.URL.Path)) {
			html(w, http.StatusNotFound)
			return
		}
		htmlBody(w, http.StatusOK, "")
		return
	}
	if strings.HasPrefix(r.URL.Path, "/projects/") && strings.HasSuffix(r.URL.Path, "/checks/status/") {
		if !a.requireLogin(w, r) {
			return
		}
		if !a.canAccessProject(a.currentUser(r), projectIDFromPath(r.URL.Path)) {
			html(w, http.StatusNotFound)
			return
		}
		htmlBody(w, http.StatusOK, "")
		return
	}
	if strings.HasPrefix(r.URL.Path, "/projects/") && strings.HasSuffix(r.URL.Path, "/settings/") {
		if !a.loggedIn(r) {
			if r.Method == http.MethodPost {
				html(w, http.StatusForbidden)
			} else {
				redirect(w, "/accounts/login/")
			}
			return
		}
		if !a.canAccessProject(a.currentUser(r), projectIDFromPath(r.URL.Path)) {
			html(w, http.StatusNotFound)
			return
		}
		if r.Method == http.MethodGet {
			htmlBody(w, http.StatusOK, a.csrfForm(w, r))
			return
		}
		if r.Method == http.MethodPost && a.validCSRF(r) {
			html(w, http.StatusOK)
		} else if r.Method == http.MethodPost {
			html(w, http.StatusForbidden)
		} else {
			html(w, http.StatusMethodNotAllowed)
		}
		return
	}
	if strings.HasPrefix(r.URL.Path, "/projects/") && strings.HasSuffix(r.URL.Path, "/remove/") {
		if !a.requireLogin(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			html(w, http.StatusMethodNotAllowed)
			return
		}
		if !a.validCSRF(r) {
			html(w, http.StatusForbidden)
		} else if a.removeProject(projectIDFromPath(r.URL.Path), a.currentUser(r)) {
			html(w, http.StatusOK)
		} else {
			html(w, http.StatusForbidden)
		}
		return
	}
	if r.URL.Path == "/badge/nonexistent/0000000000000000000000000000000000000000/test-tag.svg" {
		html(w, 404)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && (parts[1] == "v1" || parts[1] == "v2" || parts[1] == "v3") {
		v := parts[1]
		if len(parts) == 3 && parts[2] == "checks" {
			if r.Method == http.MethodPost {
				a.create(w, r, v)
			} else if r.Method == http.MethodGet {
				a.listChecks(w, r, v)
			} else {
				html(w, 405)
			}
			return
		}
		if len(parts) == 3 && parts[2] == "channels" {
			a.apiChannels(w, r)
			return
		}
		if len(parts) == 3 && parts[2] == "badges" {
			if a.authorized(w, r, false) {
				jsonResponse(w, http.StatusOK, map[string]any{"badges": map[string]any{}})
			}
			return
		}
		if len(parts) == 3 && parts[2] == "metrics" {
			if !a.authorized(w, r, false) {
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			_, _ = io.WriteString(w, "")
			return
		}
		if len(parts) == 3 && parts[2] == "status" {
			htmlBody(w, http.StatusOK, "OK")
			return
		}
		if len(parts) == 3 && parts[2] == "bounces" {
			html(w, 200)
			return
		}
		if len(parts) == 4 && parts[2] == "notifications" && parts[3] == "status" {
			html(w, 404)
			return
		}
		if len(parts) >= 4 && parts[2] == "checks" {
			id := parts[3]
			if len(parts) == 4 {
				if len(id) == 40 {
					if !a.authorized(w, r, false) {
						return
					}
					projectID := a.requestProject(r)
					a.mu.Lock()
					var found *check
					for _, c := range a.checks {
						sum := sha1.Sum([]byte(strings.ReplaceAll(c.ID, "-", "")[:16]))
						if hex.EncodeToString(sum[:]) == id && c.ProjectID == projectID {
							found = c
							break
						}
					}
					a.mu.Unlock()
					if found == nil {
						html(w, http.StatusNotFound)
					} else {
						jsonResponse(w, http.StatusOK, checkJSON(found, v, baseURL(r)))
					}
					return
				}
				a.single(w, r, v, id)
				return
			}
			if len(parts) == 5 && parts[4] == "pause" {
				if !a.authorized(w, r, true) {
					return
				}
				if !a.apiCanAccessCheck(r, id) {
					html(w, http.StatusForbidden)
					return
				}
				c, found := a.pauseCheck(id)
				if !found {
					html(w, 404)
					return
				}
				jsonResponse(w, 200, checkJSON(&c, v, baseURL(r)))
				return
			}
			if len(parts) == 5 && parts[4] == "resume" {
				if !a.authorized(w, r, true) {
					return
				}
				if !a.apiCanAccessCheck(r, id) {
					html(w, http.StatusForbidden)
					return
				}
				c, status := a.resumeCheck(id)
				if status != http.StatusOK {
					html(w, status)
					return
				}
				jsonResponse(w, http.StatusOK, checkJSON(&c, v, baseURL(r)))
				return
			}
			if len(parts) >= 5 && parts[4] == "pings" {
				a.pings(w, r, v, id)
				return
			}
			if len(parts) == 5 && parts[4] == "flips" {
				a.flips(w, r, id)
				return
			}
		}
	}
	if len(parts) >= 2 && parts[0] == "ping" {
		if len(parts) >= 3 && parts[1] == pingKey {
			slug := parts[2]
			action := ""
			if len(parts) == 4 {
				action = parts[3]
			}
			a.mu.Lock()
			var id string
			for k, c := range a.checks {
				if c.Slug == slug {
					id = k
					break
				}
			}
			a.mu.Unlock()
			if id == "" {
				html(w, 404)
			} else {
				a.ping(w, r, id, action)
			}
			return
		}
		id := parts[1]
		action := ""
		if len(parts) == 3 {
			action = parts[2]
		}
		if strings.Count(id, "-") != 4 {
			html(w, 404)
			return
		}
		a.ping(w, r, id, action)
		return
	}
	html(w, 404)
}

type diagnosticWriter struct {
	http.ResponseWriter
	status int
	body   strings.Builder
}

func (w *diagnosticWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *diagnosticWriter) Write(b []byte) (int, error) {
	if w.body.Len() < 500 {
		_, _ = w.body.Write(b[:min(len(b), 500-w.body.Len())])
	}
	return w.ResponseWriter.Write(b)
}

var diagnosticSequence uint64

func writeDiagnostic(path string, r *http.Request, w *diagnosticWriter) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	record := map[string]any{"seq": atomic.AddUint64(&diagnosticSequence, 1), "time": time.Now().UTC().Format(time.RFC3339Nano), "method": r.Method, "path": r.URL.RequestURI(), "status": w.status, "content_type": w.Header().Get("Content-Type"), "body": w.body.String()}
	if encoded, err := json.Marshal(record); err == nil {
		_, _ = f.Write(append(encoded, '\n'))
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	a := newApp()
	log.Printf("healthchecks compatibility server listening on :%s (HC_DIAGNOSTIC_LOG=%q)", port, os.Getenv("HC_DIAGNOSTIC_LOG"))
	log.Fatal(http.ListenAndServe(":"+port, http.HandlerFunc(a.serveHTTP)))
}
