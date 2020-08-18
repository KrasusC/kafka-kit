package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DataDog/kafka-kit/v3/kafkazk"
	"github.com/DataDog/kafka-kit/v3/reporter"
	"github.com/DataDog/kafka-kit/v3/reporter/datadog"
	"github.com/rcrowley/go-metrics"
)

// MetricsNamespace holds the common prefix for metric names
const MetricsNamespace = "autothrottle."

// AppMetricsOpt ...
type AppMetricsOpt struct {
	reportEnable  bool
	backend       string
	flushInterval time.Duration
	statsdAddr    string
	tags          []string
}

// initAppMetrics ...
func initAppMetrics(zk kafkazk.Handler, opt AppMetricsOpt) {
	brokerIDs, err := zk.GetBrokerIDs()
	if err != nil {
		log.Printf("Error fetching broker IDs for app metrics: %s\n", err)
	}

	for _, ID := range brokerIDs {
		for _, role := range []string{"leader", "follower"} {
			metricName := fmt.Sprintf("broker.%d.%s.%s", ID, role, "replication.throttled.rate")
			metrics.GetOrRegisterGauge(metricName, metricsRegistry)
		}
	}

	if !opt.reportEnable {
		log.Printf("Metrics reporting is disabled\n")
		return
	}

	// "broker.10004.leader.replication.throttled.rate" ->
	// name: "leader.replication.throttled.rate", tags: broker_id:10004
	transformFn := func(name string, tags []string) (string, []string) {
		parts := strings.SplitN(name, ".", 6)
		brokerID := parts[1]
		finalName := fmt.Sprintf("%s%s", MetricsNamespace, strings.Join(parts[2:], "."))
		tags = append(tags, fmt.Sprintf("broker_id:%s", brokerID))
		return finalName, tags
	}

	var opts interface{}
	switch opt.backend {
	case "datadog":
		opts = datadog.Options{
			StatsdAddr:  opt.statsdAddr,
			TransformFn: transformFn,
			Tags:        opt.tags,
		}
	}

	reporter, err := reporter.NewReporter(opt.backend, metricsRegistry, opt.flushInterval, opts)
	if err != nil {
		log.Printf("Failed to initialize the %s reporter. Metrics won't be submitted.\n", opt.backend)
		return
	}
	//	fmt.Printf("reporter: %+v\n", reporter)

	go reporter.PeriodicFlush()
}
