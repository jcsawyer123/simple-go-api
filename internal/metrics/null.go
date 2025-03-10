package metrics

import (
	"context"
	"time"
)

// NullProvider is a no-op implementation of the MetricsProvider interface
type NullProvider struct{}

// NewNullProvider creates a new null metrics provider
func NewNullProvider() *NullProvider {
	return &NullProvider{}
}

// Init is a no-op for the null provider
func (p *NullProvider) Init() error {
	return nil
}

// Close is a no-op for the null provider
func (p *NullProvider) Close() error {
	return nil
}

// Counter returns a no-op counter
func (p *NullProvider) Counter(name string, tags map[string]string) Counter {
	return &nullCounter{}
}

// Gauge returns a no-op gauge
func (p *NullProvider) Gauge(name string, tags map[string]string) Gauge {
	return &nullGauge{}
}

// Histogram returns a no-op histogram
func (p *NullProvider) Histogram(name string, tags map[string]string) Histogram {
	return &nullHistogram{}
}

// Timer returns a no-op timer
func (p *NullProvider) Timer(name string, tags map[string]string) Timer {
	return &nullTimer{}
}

// No-op implementations

type nullCounter struct{}

func (c *nullCounter) Inc()                                {}
func (c *nullCounter) Add(value float64)                   {}
func (c *nullCounter) With(tags map[string]string) Counter { return c }

type nullGauge struct{}

func (g *nullGauge) Set(value float64)                 {}
func (g *nullGauge) Inc()                              {}
func (g *nullGauge) Dec()                              {}
func (g *nullGauge) Add(value float64)                 {}
func (g *nullGauge) Sub(value float64)                 {}
func (g *nullGauge) With(tags map[string]string) Gauge { return g }

type nullHistogram struct{}

func (h *nullHistogram) Observe(value float64)                 {}
func (h *nullHistogram) With(tags map[string]string) Histogram { return h }

type nullTimer struct{}

func (t *nullTimer) Record(f func())                                                    { f() }
func (t *nullTimer) RecordWithContext(ctx context.Context, f func(ctx context.Context)) { f(ctx) }
func (t *nullTimer) Start() func()                                                      { return func() {} }
func (t *nullTimer) ObserveDuration(duration time.Duration)                             {}
func (t *nullTimer) With(tags map[string]string) Timer                                  { return t }
