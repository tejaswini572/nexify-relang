package main

import (
	"maps"
	"net/http"
	"sync"
	"time"
)

// Store-owned records and synchronized mutations live here. HTTP handlers use
// snapshots or these methods rather than retaining references into live maps.
type ping struct {
	Body, Kind, RID, Method string
	At                      time.Time
	ExitStatus              int
}

type flip struct {
	Status, Reason string
	At             time.Time
}

type channel struct {
	ID, ProjectID, Kind, Value, Name     string
	Down, Up                             bool
	URLDown, URLUp, MethodDown, MethodUp string
	HeadersDown, HeadersUp               map[string]string
	Members                              []string
	Enabled                              bool
}

type check struct {
	ID, ProjectID, Name, Slug, Tags, Description, Schedule, TZ, Status string
	Timeout, Grace                                                     int
	Methods, StartKW, SuccessKW, FailureKW                             string
	FilterBody, FilterSubject, FilterHTTPBody                          bool
	FilterDefaultFail, ManualResume                                    bool
	LastPing, LastStart                                                *time.Time
	LastStartRID                                                       string
	LastDuration                                                       *time.Duration
	Pings                                                              []ping
	Flips                                                              []flip
	ChannelIDs                                                         []string
	Sequence                                                           int
}

type app struct {
	mu        sync.Mutex
	checks    map[string]*check
	channels  []channel
	sessions  map[string]string
	project   string
	apiKeys   map[string]string
	readKeys  map[string]string
	members   map[string]map[string]bool
	projects  map[string]projectRecord
	nextCheck int
}

type projectRecord struct{ ID, Name, Owner string }

func newApp() *app {
	return &app{checks: make(map[string]*check), sessions: make(map[string]string), apiKeys: make(map[string]string), readKeys: make(map[string]string), members: make(map[string]map[string]bool), projects: make(map[string]projectRecord)}
}

func (a *app) reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checks = make(map[string]*check)
	a.channels = nil
	a.sessions = make(map[string]string)
	a.project = uuid()
	bobProject, charlieProject := uuid(), uuid()
	a.apiKeys = map[string]string{apiKey: a.project, bobAPIKey: bobProject, charlieAPIKey: charlieProject}
	a.readKeys = map[string]string{readOnlyKey: a.project}
	a.members = map[string]map[string]bool{a.project: {"alice": true, "bob": true}, bobProject: {"bob": true}, charlieProject: {"charlie": true}}
	a.projects = map[string]projectRecord{a.project: {ID: a.project, Name: "Alices Project", Owner: "alice"}, bobProject: {ID: bobProject, Name: "Bob's Project", Owner: "bob"}, charlieProject: {ID: charlieProject, Name: "Charlie's Project", Owner: "charlie"}}
	a.nextCheck = 0
}

func copyCheck(c *check) check {
	result := *c
	result.Pings = append([]ping(nil), c.Pings...)
	result.Flips = append([]flip(nil), c.Flips...)
	result.ChannelIDs = append([]string(nil), c.ChannelIDs...)
	if c.LastPing != nil {
		v := *c.LastPing
		result.LastPing = &v
	}
	if c.LastStart != nil {
		v := *c.LastStart
		result.LastStart = &v
	}
	if c.LastDuration != nil {
		v := *c.LastDuration
		result.LastDuration = &v
	}
	return result
}

func (a *app) checkSnapshot(id string) (check, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		return check{}, false
	}
	return copyCheck(c), true
}

func copyChannel(ch channel) channel {
	result := ch
	result.Members = append([]string(nil), ch.Members...)
	result.HeadersDown = maps.Clone(ch.HeadersDown)
	result.HeadersUp = maps.Clone(ch.HeadersUp)
	return result
}

func (a *app) channelSnapshot(id string) (channel, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, ch := range a.channels {
		if ch.ID == id {
			return copyChannel(ch), true
		}
	}
	return channel{}, false
}

func (a *app) pauseCheck(id string) (check, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		return check{}, false
	}
	c.Status = "paused"
	return copyCheck(c), true
}

func (a *app) resumeCheck(id string) (check, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		return check{}, http.StatusNotFound
	}
	if c.Status != "paused" {
		return check{}, http.StatusConflict
	}
	c.Status = "new"
	return copyCheck(c), http.StatusOK
}

func (a *app) updateChannelValue(id, value string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.channels {
		if a.channels[i].ID == id {
			a.channels[i].Value = value
			return true
		}
	}
	return false
}

func (a *app) renameCheck(id, name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		return false
	}
	c.Name, c.Slug = name, slugify(name)
	return true
}

func (a *app) setCheckTimeout(id string, timeout, grace int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		return false
	}
	c.Timeout, c.Grace = timeout, grace
	return true
}

func (a *app) copyCheckToProject(id string) (check, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	source := a.checks[id]
	if source == nil {
		return check{}, false
	}
	clone := copyCheck(source)
	clone.ID = uuid()
	clone.Name += " (copy)"
	clone.Slug = slugify(clone.Name)
	clone.Pings = nil
	clone.Flips = nil
	clone.LastPing = nil
	clone.LastStart = nil
	clone.LastDuration = nil
	clone.Status = "new"
	a.nextCheck++
	clone.Sequence = a.nextCheck
	a.checks[clone.ID] = &clone
	return clone, true
}

func (a *app) deleteCheck(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.checks[id] == nil {
		return false
	}
	delete(a.checks, id)
	return true
}

func (a *app) clearCheckEvents(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.checks[id]
	if c == nil {
		return false
	}
	c.Flips = nil
	return true
}

func (a *app) toggleChannel(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.channels {
		if a.channels[i].ID == id {
			a.channels[i].Enabled = !a.channels[i].Enabled
			return true
		}
	}
	return false
}

func (a *app) removeChannel(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.channels {
		if a.channels[i].ID == id {
			a.channels = append(a.channels[:i], a.channels[i+1:]...)
			for _, c := range a.checks {
				c.ChannelIDs = removeString(c.ChannelIDs, id)
			}
			return true
		}
	}
	return false
}
func removeString(items []string, target string) []string {
	out := items[:0]
	for _, item := range items {
		if item != target {
			out = append(out, item)
		}
	}
	return out
}
