package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Counter struct {
	prometheus.Counter
	*collectorMetadata
}

func NewCounter(opts prometheus.CounterOpts) *Counter {
	return &Counter{
		Counter:           prometheus.NewCounter(opts),
		collectorMetadata: NewCollectorMetadata(),
	}
}

func (c *Counter) Inc() {
	c.touch()
	c.Counter.Inc()
}

func (c *Counter) Add(v float64) {
	c.touch()
	c.Counter.Add(v)
}

type CounterVec struct {
	*collectorMetadata
	*prometheus.CounterVec
}

func NewCounterVec(opts prometheus.CounterOpts, labelNames []string) *CounterVec {
	return &CounterVec{
		collectorMetadata: NewCollectorMetadata(),
		CounterVec:        prometheus.NewCounterVec(opts, labelNames),
	}
}

func (c *CounterVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Counter, error) {
	c.touch()
	return c.CounterVec.GetMetricWithLabelValues(lvs...)
}

func (c *CounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	c.touch()
	return c.CounterVec.WithLabelValues(lvs...)
}

func (c *CounterVec) GetMetricWith(labels prometheus.Labels) (prometheus.Counter, error) {
	c.touch()
	return c.CounterVec.GetMetricWith(labels)
}

func (c *CounterVec) With(labels prometheus.Labels) prometheus.Counter {
	c.touch()
	return c.CounterVec.With(labels)
}

func (c *CounterVec) CurryWith(labels prometheus.Labels) (*prometheus.CounterVec, error) {
	c.touch()
	return c.CounterVec.CurryWith(labels)
}

func (c *CounterVec) MustCurryWith(labels prometheus.Labels) *prometheus.CounterVec {
	c.touch()
	return c.CounterVec.MustCurryWith(labels)
}
