package utils

import "sync"

// ExecutionContext keeps shared state such as extracted variables
// between declarative test steps and suites.
type ExecutionContext struct {
	Vars map[string]string
	mu   sync.RWMutex
}

func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{Vars: map[string]string{}}
}

func (c *ExecutionContext) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Vars[key] = value
}

func (c *ExecutionContext) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.Vars[key]
	return v, ok
}

// Snapshot returns a copy of the variables map for safe concurrent use.
func (c *ExecutionContext) Snapshot() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]string, len(c.Vars))
	for k, v := range c.Vars {
		out[k] = v
	}
	return out
}
