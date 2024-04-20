package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Gauge struct {
	*collectorMetadata
	prometheus.Gauge
}

type GaugeOpts prometheus.GaugeOpts

func NewGauge(opts prometheus.GaugeOpts) *Gauge {
	return &Gauge{
		collectorMetadata: NewCollectorMetadata(),
		Gauge:             prometheus.NewGauge(opts),
	}
}

func (g *Gauge) Add(v float64) {
	g.touch()
	g.Gauge.Add(v)
}
func (g *Gauge) Set(v float64) {
	g.touch()
	g.Gauge.Set(v)
}
func (g *Gauge) Inc() {
	g.touch()
	g.Gauge.Inc()
}

func (g *Gauge) Dec() {
	g.touch()
	g.Gauge.Dec()
}

func (g *Gauge) Sub(v float64) {
	g.touch()
	g.Gauge.Sub(v)
}

type GaugeVec struct {
	*collectorMetadata
	*prometheus.GaugeVec
}

func NewGaugeVec(opts prometheus.GaugeOpts, labels []string) *GaugeVec {
	return &GaugeVec{
		collectorMetadata: NewCollectorMetadata(),
		GaugeVec:          prometheus.NewGaugeVec(opts, labels),
	}
}

func (v *GaugeVec) CurryWith(labels prometheus.Labels) (*prometheus.GaugeVec, error) {
	v.touch()
	return v.GaugeVec.CurryWith(labels)
}
func (v *GaugeVec) GetMetricWith(labels prometheus.Labels) (prometheus.Gauge, error) {
	v.touch()
	return v.GaugeVec.GetMetricWith(labels)
}
func (v *GaugeVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Gauge, error) {
	v.touch()
	return v.GaugeVec.GetMetricWithLabelValues(lvs...)
}
func (v *GaugeVec) MustCurryWith(labels prometheus.Labels) *prometheus.GaugeVec {
	v.touch()
	return v.GaugeVec.MustCurryWith(labels)
}
func (v *GaugeVec) With(labels prometheus.Labels) prometheus.Gauge {
	v.touch()
	return v.GaugeVec.With(labels)
}
func (v *GaugeVec) WithLabelValues(lvs ...string) prometheus.Gauge {
	v.touch()
	return v.GaugeVec.WithLabelValues(lvs...)
}
