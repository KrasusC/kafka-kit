package reporter

import (
	"fmt"
	"time"

	"github.com/DataDog/kafka-kit/v3/reporter/datadog"
	"github.com/rcrowley/go-metrics"
)

// Reporter provides the main flush logic to submit metrics to a
// specific metrics backend
type Reporter interface {
	Flush()
	PeriodicFlush()
}

// NewReporter ...
func NewReporter(backend string, registry metrics.Registry, interval time.Duration, options interface{}) (Reporter, error) {
	switch backend {
	case "datadog":
		opts, ok := options.(datadog.Options)
		if !ok {
			return nil, fmt.Errorf("Invalid datadog options %q", options)
		}
		return datadog.NewReporter(registry, interval, opts)
	default:
		return nil, fmt.Errorf("Unknown metrics backend: %s", backend)
	}
}
