package metrics

import (
	"context"
	"time"
)

// MetricsProvider defines the interface for metrics providers
type MetricsProvider interface {
	// Counter methods
	Counter(name string, tags map[string]string) Counter

	// Gauge methods
	Gauge(name string, tags map[string]string) Gauge

	// Histogram methods
	Histogram(name string, tags map[string]string) Histogram

	// Timer convenience methods
	Timer(name string, tags map[string]string) Timer

	// Initialize the metrics provider
	Init() error

	// Close and flush any remaining metrics
	Close() error
}

// Counter represents a metric that accumulates values
type Counter interface {
	// Inc increments the counter by 1
	Inc()

	// Add adds the given value to the counter
	Add(value float64)

	// With returns a new Counter with added tags
	With(tags map[string]string) Counter
}

// Gauge represents a metric that can arbitrarily go up and down
type Gauge interface {
	// Set sets the gauge to an arbitrary value
	Set(value float64)

	// Inc increments the gauge by 1
	Inc()

	// Dec decrements the gauge by 1
	Dec()

	// Add adds the given value to the gauge
	Add(value float64)

	// Sub subtracts the given value from the gauge
	Sub(value float64)

	// With returns a new Gauge with added tags
	With(tags map[string]string) Gauge
}

// Histogram represents a metric that samples observations
type Histogram interface {
	// Observe adds a single observation to the histogram
	Observe(value float64)

	// With returns a new Histogram with added tags
	With(tags map[string]string) Histogram
}

// Timer is a convenience interface for timing operations
type Timer interface {
	// Record records the duration of the given function
	Record(f func())

	// RecordWithContext records the duration of a function that accepts a context
	RecordWithContext(ctx context.Context, f func(ctx context.Context))

	// Start starts the timer and returns a function to stop and record the duration
	Start() func()

	// With returns a new Timer with added tags
	With(tags map[string]string) Timer
}

// Reporter is the main metrics reporting interface that can report to multiple providers
type Reporter interface {
	MetricsProvider

	// AddProvider adds a new metrics provider
	AddProvider(provider MetricsProvider)

	// RemoveProvider removes a metrics provider
	RemoveProvider(provider MetricsProvider)
}

// DefaultReporter is the standard implementation of Reporter that reports to multiple providers
type DefaultReporter struct {
	providers []MetricsProvider
}

// NewReporter creates a new DefaultReporter instance
func NewReporter(providers ...MetricsProvider) *DefaultReporter {
	return &DefaultReporter{
		providers: providers,
	}
}

// AddProvider adds a new metrics provider
func (r *DefaultReporter) AddProvider(provider MetricsProvider) {
	r.providers = append(r.providers, provider)
}

// RemoveProvider removes a metrics provider
func (r *DefaultReporter) RemoveProvider(provider MetricsProvider) {
	for i, p := range r.providers {
		if p == provider {
			r.providers = append(r.providers[:i], r.providers[i+1:]...)
			return
		}
	}
}

// Counter implementation for DefaultReporter
func (r *DefaultReporter) Counter(name string, tags map[string]string) Counter {
	counters := make([]Counter, 0, len(r.providers))
	for _, p := range r.providers {
		counters = append(counters, p.Counter(name, tags))
	}
	return &multiCounter{counters: counters}
}

// Gauge implementation for DefaultReporter
func (r *DefaultReporter) Gauge(name string, tags map[string]string) Gauge {
	gauges := make([]Gauge, 0, len(r.providers))
	for _, p := range r.providers {
		gauges = append(gauges, p.Gauge(name, tags))
	}
	return &multiGauge{gauges: gauges}
}

// Histogram implementation for DefaultReporter
func (r *DefaultReporter) Histogram(name string, tags map[string]string) Histogram {
	histograms := make([]Histogram, 0, len(r.providers))
	for _, p := range r.providers {
		histograms = append(histograms, p.Histogram(name, tags))
	}
	return &multiHistogram{histograms: histograms}
}

// Timer implementation for DefaultReporter
func (r *DefaultReporter) Timer(name string, tags map[string]string) Timer {
	timers := make([]Timer, 0, len(r.providers))
	for _, p := range r.providers {
		timers = append(timers, p.Timer(name, tags))
	}
	return &multiTimer{timers: timers}
}

// Init initializes all providers
func (r *DefaultReporter) Init() error {
	for _, p := range r.providers {
		if err := p.Init(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes all providers
func (r *DefaultReporter) Close() error {
	var lastErr error
	for _, p := range r.providers {
		if err := p.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Multi-provider implementations

type multiCounter struct {
	counters []Counter
}

func (m *multiCounter) Inc() {
	for _, c := range m.counters {
		c.Inc()
	}
}

func (m *multiCounter) Add(value float64) {
	for _, c := range m.counters {
		c.Add(value)
	}
}

func (m *multiCounter) With(tags map[string]string) Counter {
	counters := make([]Counter, 0, len(m.counters))
	for _, c := range m.counters {
		counters = append(counters, c.With(tags))
	}
	return &multiCounter{counters: counters}
}

type multiGauge struct {
	gauges []Gauge
}

func (m *multiGauge) Set(value float64) {
	for _, g := range m.gauges {
		g.Set(value)
	}
}

func (m *multiGauge) Inc() {
	for _, g := range m.gauges {
		g.Inc()
	}
}

func (m *multiGauge) Dec() {
	for _, g := range m.gauges {
		g.Dec()
	}
}

func (m *multiGauge) Add(value float64) {
	for _, g := range m.gauges {
		g.Add(value)
	}
}

func (m *multiGauge) Sub(value float64) {
	for _, g := range m.gauges {
		g.Sub(value)
	}
}

func (m *multiGauge) With(tags map[string]string) Gauge {
	gauges := make([]Gauge, 0, len(m.gauges))
	for _, g := range m.gauges {
		gauges = append(gauges, g.With(tags))
	}
	return &multiGauge{gauges: gauges}
}

type multiHistogram struct {
	histograms []Histogram
}

func (m *multiHistogram) Observe(value float64) {
	for _, h := range m.histograms {
		h.Observe(value)
	}
}

func (m *multiHistogram) With(tags map[string]string) Histogram {
	histograms := make([]Histogram, 0, len(m.histograms))
	for _, h := range m.histograms {
		histograms = append(histograms, h.With(tags))
	}
	return &multiHistogram{histograms: histograms}
}

type multiTimer struct {
	timers []Timer
}

func (m *multiTimer) Record(f func()) {
	start := time.Now()
	f()
	duration := time.Since(start)

	// Record the same duration in all timers
	for _, t := range m.timers {
		t.(interface{ ObserveDuration(time.Duration) }).ObserveDuration(duration)
	}
}

func (m *multiTimer) RecordWithContext(ctx context.Context, f func(ctx context.Context)) {
	start := time.Now()
	f(ctx)
	duration := time.Since(start)

	// Record the same duration in all timers
	for _, t := range m.timers {
		t.(interface{ ObserveDuration(time.Duration) }).ObserveDuration(duration)
	}
}

func (m *multiTimer) Start() func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		for _, t := range m.timers {
			t.(interface{ ObserveDuration(time.Duration) }).ObserveDuration(duration)
		}
	}
}

func (m *multiTimer) With(tags map[string]string) Timer {
	timers := make([]Timer, 0, len(m.timers))
	for _, t := range m.timers {
		timers = append(timers, t.With(tags))
	}
	return &multiTimer{timers: timers}
}

// Global metrics reporter
var globalReporter Reporter = NewReporter()

// InitGlobal initializes the global reporter with the given providers
func InitGlobal(providers ...MetricsProvider) error {
	for _, provider := range providers {
		globalReporter.AddProvider(provider)
	}
	return globalReporter.Init()
}

// Global helper functions

// Counter returns a counter from the global reporter
func CounterMetric(name string, tags map[string]string) Counter {
	return globalReporter.Counter(name, tags)
}

// Gauge returns a gauge from the global reporter
func GaugeMetric(name string, tags map[string]string) Gauge {
	return globalReporter.Gauge(name, tags)
}

// Histogram returns a histogram from the global reporter
func HistogramMetric(name string, tags map[string]string) Histogram {
	return globalReporter.Histogram(name, tags)
}

// Timer returns a timer from the global reporter
func TimerMetric(name string, tags map[string]string) Timer {
	return globalReporter.Timer(name, tags)
}

// CloseGlobal closes the global reporter
func CloseGlobal() error {
	return globalReporter.Close()
}
