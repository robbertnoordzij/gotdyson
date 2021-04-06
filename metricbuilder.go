package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "gotdyson"

var (
	defaultLabels = []string{"device"}
)

type MetricBuilder struct {
	ValueType  prometheus.ValueType
	Desc       *prometheus.Desc
	MetricFunc func(value float64) (prometheus.Metric, error)
	Extractor  func(frame *Frame) float64
}

func (mb MetricBuilder) String() string {
	return mb.Desc.String()
}

var metricBuilders = map[string]MetricBuilder{
	"temperature": MetricBuilder{
		ValueType: prometheus.GaugeValue,
		Desc: prometheus.NewDesc(
			namespace+"_temperature_k",
			"Temperature reading from Dyson in Kelvin",
			defaultLabels,
			prometheus.Labels{},
		),
		Extractor: func(frame *Frame) float64 {
			return frame.temperature
		},
	},
	"dust": MetricBuilder{
		ValueType: prometheus.GaugeValue,
		Desc: prometheus.NewDesc(
			namespace+"_dust",
			"Dust reading from Dyson",
			defaultLabels,
			prometheus.Labels{},
		),
		Extractor: func(frame *Frame) float64 {
			return frame.dust
		},
	},
	"volatileCompounds": MetricBuilder{
		ValueType: prometheus.GaugeValue,
		Desc: prometheus.NewDesc(
			namespace+"_volatile_organic_compounds_ug_ft3",
			"Volatile organic compounds reading from Dyson in micrograms per foot cubed",
			defaultLabels,
			prometheus.Labels{},
		),
		Extractor: func(frame *Frame) float64 {
			return frame.volatileOrganicCompounds
		},
	},
	"humidity": MetricBuilder{
		ValueType: prometheus.GaugeValue,
		Desc: prometheus.NewDesc(
			namespace+"_humidity_p",
			"Humidity reading from Dyson in percent",
			defaultLabels,
			prometheus.Labels{},
		),
		Extractor: func(frame *Frame) float64 {
			return frame.humidity
		},
	},
}