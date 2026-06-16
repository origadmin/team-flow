// Package templates embeds the tool-level sub_flow templates that ship
// with the binary. Templates are read-only — once embedded, they are part
// of the tool's identity. A team can reference a template by ID in
// team.json (e.g. {"id":"bugfix-flow"} with no "file" key); that creates
// a child flow that inherits the template verbatim.
//
// If a team adds a "file" key, the JSON on disk is loaded and treated as
// a fork — no longer bound to the template.
package templates

import (
	"embed"
	"fmt"
	"strings"
	"sync"
)

//go:embed all:content
var content embed.FS

var (
	cache  = map[string][]byte{}
	loaded = false
	mu     sync.RWMutex
)

// Reset clears the in-memory cache. Test-only.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	cache = map[string][]byte{}
	loaded = false
}

func loadAll() {
	mu.Lock()
	defer mu.Unlock()
	if loaded {
		return
	}
	entries, err := content.ReadDir("content")
	if err != nil {
		mu.Unlock()
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := content.ReadFile("content/" + e.Name())
		if err != nil {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		cache[id] = data
	}
	loaded = true
}

// Has reports whether a template with the given id is embedded.
func Has(id string) bool {
	loadAll()
	mu.RLock()
	defer mu.RUnlock()
	_, ok := cache[id]
	return ok
}

// Get returns the raw JSON bytes of the named template.
// Returns an error if the template is not embedded.
func Get(id string) ([]byte, error) {
	loadAll()
	mu.RLock()
	defer mu.RUnlock()
	if data, ok := cache[id]; ok {
		// return a copy to prevent caller mutation of the cache
		out := make([]byte, len(data))
		copy(out, data)
		return out, nil
	}
	return nil, fmt.Errorf("template %q is not embedded. Available: %v", id, List())
}

// List returns all embedded template IDs.
func List() []string {
	loadAll()
	mu.RLock()
	defer mu.RUnlock()
	ids := make([]string, 0, len(cache))
	for id := range cache {
		ids = append(ids, id)
	}
	return ids
}

// SentinelPath returns a synthetic path that the flow resolver recognises
// as "this id is a template, not a disk file". Use LoadBySentinel to parse
// the bytes for that id.
func SentinelPath(id string) string {
	return "embed://templates/" + id + ".json"
}

// IsSentinel reports whether the given path is a template sentinel.
func IsSentinel(path string) bool {
	return strings.HasPrefix(path, "embed://templates/")
}

// IDFromSentinel extracts the template id from a sentinel path.
func IDFromSentinel(path string) (string, bool) {
	if !IsSentinel(path) {
		return "", false
	}
	rest := strings.TrimPrefix(path, "embed://templates/")
	rest = strings.TrimSuffix(rest, ".json")
	return rest, true
}
