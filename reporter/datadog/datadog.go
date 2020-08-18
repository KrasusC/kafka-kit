package datadog

import (
	"fmt"
	"log"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/rcrowley/go-metrics"
)

var (
	defaultStatsdAddr = "localhost:8125"
)

// return actual names
type TransformFn func(string, []string) (string, []string)

// Options ..
type Options struct {
	StatsdAddr  string
	Tags        []string
	TransformFn TransformFn
}

// Reporter ...
type Reporter struct {
	registry    metrics.Registry
	interval    time.Duration
	client      *statsd.Client
	tags        []string
	transformFn TransformFn
}

// NewReporter creates a new Reporter.
func NewReporter(registry metrics.Registry, interval time.Duration, options Options) (*Reporter, error) {
	var statsdAddr string
	if options.StatsdAddr != "" {
		statsdAddr = options.StatsdAddr
	} else {
		statsdAddr = defaultStatsdAddr
	}

	fmt.Printf("statsdAddr: %v\n", statsdAddr)

	client, err := statsd.New(statsdAddr)
	if err != nil {
		return nil, err
	}

	return &Reporter{
		registry:    registry,
		interval:    interval,
		client:      client,
		tags:        options.Tags,
		transformFn: options.TransformFn,
	}, nil
}

// PeriodicFlush calls Flush periodically according to the given interval
func (r *Reporter) PeriodicFlush() {
	for range time.Tick(r.interval) {
		r.Flush()
	}
}

// Flush submits a snapshot of the registry to the StatsD client (which sends it to Datadog)
func (r *Reporter) Flush() {
	r.registry.Each(func(metricName string, i interface{}) {
		fmt.Printf("registry:flush:name: %v == i: %v\n", metricName, i)
		name := metricName
		tags := r.tags
		if r.transformFn != nil {
			name, tags = r.transformFn(metricName, r.tags)
		}
		fmt.Printf("name: %v, tags: %v\n", name, tags)
		switch metric := i.(type) {
		case metrics.Gauge:
			err := r.client.Gauge(name, float64(metric.Value()), tags, 1)
			if err != nil {
				log.Printf("Failed to send metrics via statsd\n")
			}
		default:
			// Ignore other types. This will be added when needed.
		}
	})
}
