package suite

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const excludedPrefix = "dolchevideo-"

// Loader discovers microservices and test suites.
type Loader struct {
	WorkDir string
}

func NewLoader(workDir string) *Loader {
	return &Loader{WorkDir: workDir}
}

// DiscoverServices scans the work directory and returns services that should
// be part of the test coverage (i.e., excluding dolchevideo-*).
func (l *Loader) DiscoverServices() ([]string, error) {
	entries, err := os.ReadDir(l.WorkDir)
	if err != nil {
		return nil, err
	}
	services := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(strings.ToLower(name), excludedPrefix) {
			continue
		}
		services = append(services, name)
	}
	if len(services) == 0 {
		return nil, errors.New("no services discovered; ensure /work contains microservices")
	}
	return services, nil
}

// LoadTestSuites returns all registered suites. Suites register themselves via
// init() functions in their respective packages.
func (l *Loader) LoadTestSuites() []*TestSuite {
	suites := RegisteredSuites()
	services, svcErr := l.DiscoverServices()
	for _, ts := range suites {
		if len(ts.Services) == 0 {
			if svcErr == nil {
				ts.Services = services
			}
		}
		ts.Config = mergeMaps(map[string]any{"workdir": filepath.Clean(l.WorkDir)}, ts.Config)
	}
	return suites
}

func mergeMaps(base map[string]any, overrides map[string]any) map[string]any {
	if overrides == nil {
		return base
	}
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	return out
}
