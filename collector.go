package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"sync"
)

type DysonCollector struct {
	sync.Mutex
	metrics []prometheus.Metric
}

func (dc *DysonCollector) Collect(ch chan<- prometheus.Metric) {
	dc.Lock()
	defer dc.Unlock()
	for _, m := range dc.metrics {
		ch <- m
	}
}

func (dc *DysonCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, mb := range metricBuilders {
		ch <- mb.Desc
	}
}

func (dc *DysonCollector) Update(f Frame) {
	var metrics []prometheus.Metric
	for _, mb := range metricBuilders {
		value := mb.Extractor(&f)

		m, err := prometheus.NewConstMetric(
			mb.Desc,
			mb.ValueType,
			value,
			f.deviceName,
		)
		metrics = append(metrics, m)

		if err != nil {
			log.Printf("Could not create prometheus metric for %s\n", mb)
			continue
		}
	}
	dc.Lock()
	defer dc.Unlock()
	dc.metrics = metrics
}