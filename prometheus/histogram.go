package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Histogram struct {
	*collectorMetadata
	prometheus.Histogram
}

type HistogramOpts prometheus.HistogramOpts

func NewHistogram(opts prometheus.HistogramOpts) *Histogram {
	return &Histogram{
		collectorMetadata: NewCollectorMetadata(),
		Histogram:         prometheus.NewHistogram(opts),
	}
}

func (h *Histogram) Observe(v float64) {
	h.touch()
	h.Histogram.Observe(v)
}

type HistogramVec struct {
	*collectorMetadata
	*prometheus.HistogramVec
}

func NewHistogramVec(opts prometheus.HistogramOpts, labels []string) *HistogramVec {
	return &HistogramVec{
		collectorMetadata: NewCollectorMetadata(),
		HistogramVec:      prometheus.NewHistogramVec(opts, labels),
	}
}

func (v *HistogramVec) CurryWith(labels prometheus.Labels) (prometheus.ObserverVec, error) {
	v.touch()
	return v.HistogramVec.CurryWith(labels)
}
func (v *HistogramVec) GetMetricWith(labels prometheus.Labels) (prometheus.Observer, error) {
	v.touch()
	return v.HistogramVec.GetMetricWith(labels)
}
func (v *HistogramVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Observer, error) {
	v.touch()
	return v.HistogramVec.GetMetricWithLabelValues(lvs...)
}
func (v *HistogramVec) MustCurryWith(labels prometheus.Labels) prometheus.ObserverVec {
	v.touch()
	return v.HistogramVec.MustCurryWith(labels)
}
func (v *HistogramVec) With(labels prometheus.Labels) prometheus.Observer {
	v.touch()
	return v.HistogramVec.With(labels)
}
func (v *HistogramVec) WithLabelValues(lvs ...string) prometheus.Observer {
	v.touch()
	return v.HistogramVec.WithLabelValues(lvs...)
}
