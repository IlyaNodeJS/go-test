package suite

import "sync"

var (
	registryMu sync.RWMutex
	registry   []*TestSuite
)

// RegisterSuite is called by suite definition packages to make them available
// to the runner without manual wiring.
func RegisterSuite(ts *TestSuite) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(registry, ts)
}

// RegisteredSuites returns a snapshot copy so callers can mutate safely.
func RegisteredSuites() []*TestSuite {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]*TestSuite, len(registry))
	copy(out, registry)
	return out
}
