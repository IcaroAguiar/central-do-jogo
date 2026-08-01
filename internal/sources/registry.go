package sources

import (
	"fmt"
	"sync"
)

// Registry holds registered adapters indexed by source ID.
type Registry struct {
	mu        sync.RWMutex
	adapters  map[string]Adapter
	manifests map[string]*Manifest
}

// NewRegistry creates an empty adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters:  make(map[string]Adapter),
		manifests: make(map[string]*Manifest),
	}
}

// Register adds an adapter with its manifest. It rejects adapters without a
// valid manifest or with a source ID mismatch.
func (r *Registry) Register(adapter Adapter, manifest *Manifest) error {
	if manifest == nil {
		return fmt.Errorf("cannot register adapter %q: manifest is nil", adapter.SourceID())
	}
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("cannot register adapter %q: %w", adapter.SourceID(), err)
	}
	if adapter.SourceID() != manifest.SourceID {
		return fmt.Errorf("adapter source ID %q does not match manifest source ID %q",
			adapter.SourceID(), manifest.SourceID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[adapter.SourceID()]; exists {
		return fmt.Errorf("adapter %q is already registered", adapter.SourceID())
	}

	r.adapters[adapter.SourceID()] = adapter
	r.manifests[adapter.SourceID()] = manifest
	return nil
}

// Lookup returns the adapter for the given source ID, or nil if not found.
func (r *Registry) Lookup(sourceID string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[sourceID]
	return a, ok
}

// Manifest returns the manifest for the given source ID.
func (r *Registry) Manifest(sourceID string) (*Manifest, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.manifests[sourceID]
	return m, ok
}

// All returns all registered source IDs.
func (r *Registry) All() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.adapters))
	for id := range r.adapters {
		ids = append(ids, id)
	}
	return ids
}
