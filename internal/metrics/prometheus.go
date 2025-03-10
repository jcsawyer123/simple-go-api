package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusProvider implements MetricsProvider for Prometheus
type PrometheusProvider struct {
	namespace string
	subsystem string
	registry  *prometheus.Registry
}

// PrometheusConfig contains configuration for the Prometheus provider
type PrometheusConfig struct {
	// Namespace for metrics (prefix)
	Namespace string
	// Subsystem for metrics (secondary prefix)
	Subsystem string
	// Registry is an optional custom Prometheus registry
	Registry *prometheus.Registry
}

// NewPrometheusProvider creates a new Prometheus metrics provider
func NewPrometheusProvider(config PrometheusConfig) *PrometheusProvider {
	return &PrometheusProvider{
		namespace: config.Namespace,
		subsystem: config.Subsystem,
		registry:  prometheus.DefaultRegisterer.(*prometheus.Registry),
	}
}

// Init initializes the Prometheus provider
func (p *PrometheusProvider) Init() error {
	return nil
}

// Close closes the Prometheus provider
func (p *PrometheusProvider) Close() error {
	return nil
}

// Counter returns a new Prometheus counter
func (p *PrometheusProvider) Counter(name string, tags map[string]string) Counter {
	// If no tags, use a simple counter rather than a CounterVec
	if len(tags) == 0 {
		counter := promauto.With(p.registry).NewCounter(prometheus.CounterOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      name, // Basic help text
		})
		return &prometheusSimpleCounter{counter: counter}
	}

	labels, labelValues := tagsToLabelValues(tags)

	opts := prometheus.CounterOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      name, // Basic help text
	}

	counterVec := promauto.With(p.registry).NewCounterVec(opts, labels)
	return &prometheusCounter{
		counterVec:  counterVec,
		counter:     counterVec.WithLabelValues(labelValues...),
		labelNames:  labels,
		labelValues: labelValues,
	}
}

// Gauge returns a new Prometheus gauge
func (p *PrometheusProvider) Gauge(name string, tags map[string]string) Gauge {
	// If no tags, use a simple gauge rather than a GaugeVec
	if len(tags) == 0 {
		gauge := promauto.With(p.registry).NewGauge(prometheus.GaugeOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      name, // Basic help text
		})
		return &prometheusSimpleGauge{gauge: gauge}
	}

	labels, labelValues := tagsToLabelValues(tags)

	opts := prometheus.GaugeOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      name, // Basic help text
	}

	gaugeVec := promauto.With(p.registry).NewGaugeVec(opts, labels)
	return &prometheusGauge{
		gaugeVec:    gaugeVec,
		gauge:       gaugeVec.WithLabelValues(labelValues...),
		labelNames:  labels,
		labelValues: labelValues,
	}
}

// Histogram returns a new Prometheus histogram
func (p *PrometheusProvider) Histogram(name string, tags map[string]string) Histogram {
	// If no tags, use a simple histogram rather than a HistogramVec
	if len(tags) == 0 {
		histogram := promauto.With(p.registry).NewHistogram(prometheus.HistogramOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      name, // Basic help text
			// Default buckets, can be customized
			Buckets: prometheus.DefBuckets,
		})
		return &prometheusSimpleHistogram{histogram: histogram}
	}

	labels, labelValues := tagsToLabelValues(tags)

	opts := prometheus.HistogramOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      name, // Basic help text
		// Default buckets, can be customized
		Buckets: prometheus.DefBuckets,
	}

	histogramVec := promauto.With(p.registry).NewHistogramVec(opts, labels)
	return &prometheusHistogram{
		histogramVec: histogramVec,
		histogram:    histogramVec.WithLabelValues(labelValues...),
		labelNames:   labels,
		labelValues:  labelValues,
	}
}

// Timer returns a new Prometheus timer
func (p *PrometheusProvider) Timer(name string, tags map[string]string) Timer {
	return &prometheusTimer{
		histogram: p.Histogram(name+"_duration_seconds", tags),
	}
}

// Implementation types for Prometheus

// Simple counter (no labels)
type prometheusSimpleCounter struct {
	counter prometheus.Counter
}

func (c *prometheusSimpleCounter) Inc() {
	c.counter.Inc()
}

func (c *prometheusSimpleCounter) Add(value float64) {
	c.counter.Add(value)
}

func (c *prometheusSimpleCounter) With(tags map[string]string) Counter {
	// Since this is a simple counter with no labels, we can't add labels later
	// The best we can do is just return the same counter
	// A more robust implementation would create a new CounterVec at this point
	// but that would be complex and potentially lead to duplicate metrics
	return c
}

type prometheusCounter struct {
	counterVec  *prometheus.CounterVec
	counter     prometheus.Counter
	labelNames  []string
	labelValues []string
}

func (c *prometheusCounter) Inc() {
	c.counter.Inc()
}

func (c *prometheusCounter) Add(value float64) {
	c.counter.Add(value)
}

func (c *prometheusCounter) With(tags map[string]string) Counter {
	mergedTags := mergeTags(c.labelNames, c.labelValues, tags)

	labels, labelValues := tagsToLabelValues(mergedTags)
	return &prometheusCounter{
		counterVec:  c.counterVec,
		counter:     c.counterVec.WithLabelValues(labelValues...),
		labelNames:  labels,
		labelValues: labelValues,
	}
}

// Simple gauge (no labels)
type prometheusSimpleGauge struct {
	gauge prometheus.Gauge
}

func (g *prometheusSimpleGauge) Set(value float64) {
	g.gauge.Set(value)
}

func (g *prometheusSimpleGauge) Inc() {
	g.gauge.Inc()
}

func (g *prometheusSimpleGauge) Dec() {
	g.gauge.Dec()
}

func (g *prometheusSimpleGauge) Add(value float64) {
	g.gauge.Add(value)
}

func (g *prometheusSimpleGauge) Sub(value float64) {
	g.gauge.Sub(value)
}

func (g *prometheusSimpleGauge) With(tags map[string]string) Gauge {
	// Since this is a simple gauge with no labels, we can't add labels later
	return g
}

type prometheusGauge struct {
	gaugeVec    *prometheus.GaugeVec
	gauge       prometheus.Gauge
	labelNames  []string
	labelValues []string
}

func (g *prometheusGauge) Set(value float64) {
	g.gauge.Set(value)
}

func (g *prometheusGauge) Inc() {
	g.gauge.Inc()
}

func (g *prometheusGauge) Dec() {
	g.gauge.Dec()
}

func (g *prometheusGauge) Add(value float64) {
	g.gauge.Add(value)
}

func (g *prometheusGauge) Sub(value float64) {
	g.gauge.Sub(value)
}

func (g *prometheusGauge) With(tags map[string]string) Gauge {
	mergedTags := mergeTags(g.labelNames, g.labelValues, tags)

	labels, labelValues := tagsToLabelValues(mergedTags)
	return &prometheusGauge{
		gaugeVec:    g.gaugeVec,
		gauge:       g.gaugeVec.WithLabelValues(labelValues...),
		labelNames:  labels,
		labelValues: labelValues,
	}
}

// Simple histogram (no labels)
type prometheusSimpleHistogram struct {
	histogram prometheus.Histogram
}

func (h *prometheusSimpleHistogram) Observe(value float64) {
	h.histogram.Observe(value)
}

func (h *prometheusSimpleHistogram) With(tags map[string]string) Histogram {
	// Since this is a simple histogram with no labels, we can't add labels later
	return h
}

type prometheusHistogram struct {
	histogramVec *prometheus.HistogramVec
	histogram    prometheus.Observer
	labelNames   []string
	labelValues  []string
}

func (h *prometheusHistogram) Observe(value float64) {
	h.histogram.Observe(value)
}

func (h *prometheusHistogram) With(tags map[string]string) Histogram {
	mergedTags := mergeTags(h.labelNames, h.labelValues, tags)

	labels, labelValues := tagsToLabelValues(mergedTags)
	return &prometheusHistogram{
		histogramVec: h.histogramVec,
		histogram:    h.histogramVec.WithLabelValues(labelValues...),
		labelNames:   labels,
		labelValues:  labelValues,
	}
}

type prometheusTimer struct {
	histogram Histogram
}

func (t *prometheusTimer) Record(f func()) {
	start := time.Now()
	f()
	duration := time.Since(start)
	t.ObserveDuration(duration)
}

func (t *prometheusTimer) RecordWithContext(ctx context.Context, f func(ctx context.Context)) {
	start := time.Now()
	f(ctx)
	duration := time.Since(start)
	t.ObserveDuration(duration)
}

func (t *prometheusTimer) Start() func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		t.ObserveDuration(duration)
	}
}

func (t *prometheusTimer) ObserveDuration(duration time.Duration) {
	t.histogram.Observe(duration.Seconds())
}

func (t *prometheusTimer) With(tags map[string]string) Timer {
	return &prometheusTimer{
		histogram: t.histogram.With(tags),
	}
}

// Helper functions

// tagsToLabelValues converts a map of tags to prometheus label names and values
func tagsToLabelValues(tags map[string]string) ([]string, []string) {
	labelNames := make([]string, 0, len(tags))
	labelValues := make([]string, 0, len(tags))

	for name, value := range tags {
		labelNames = append(labelNames, name)
		labelValues = append(labelValues, value)
	}

	return labelNames, labelValues
}

// mergeTags merges existing labels with new tags
func mergeTags(labelNames []string, labelValues []string, tags map[string]string) map[string]string {
	result := make(map[string]string, len(labelNames))

	// Copy existing labels
	for i, name := range labelNames {
		result[name] = labelValues[i]
	}

	// Add new tags, overwriting if necessary
	for name, value := range tags {
		result[name] = value
	}

	return result
}
