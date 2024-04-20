package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Summary struct {
	*collectorMetadata
	prometheus.Summary
}

type SummaryOpts prometheus.SummaryOpts

func NewSummary(opts prometheus.SummaryOpts) *Summary {
	return &Summary{
		collectorMetadata: NewCollectorMetadata(),
		Summary:           prometheus.NewSummary(opts),
	}
}

func (s *Summary) Observe(v float64) {
	s.touch()
	s.Summary.Observe(v)
}

type SummaryVec struct {
	*collectorMetadata
	*prometheus.SummaryVec
}

func NewSummaryVec(opts prometheus.SummaryOpts, labels []string) *SummaryVec {
	return &SummaryVec{
		collectorMetadata: NewCollectorMetadata(),
		SummaryVec:        prometheus.NewSummaryVec(opts, labels),
	}
}

func (v *SummaryVec) CurryWith(labels prometheus.Labels) (prometheus.ObserverVec, error) {
	v.touch()
	return v.SummaryVec.CurryWith(labels)
}
func (v *SummaryVec) GetMetricWith(labels prometheus.Labels) (prometheus.Observer, error) {
	v.touch()
	return v.SummaryVec.GetMetricWith(labels)
}
func (v *SummaryVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Observer, error) {
	v.touch()
	return v.SummaryVec.GetMetricWithLabelValues(lvs...)
}
func (v *SummaryVec) MustCurryWith(labels prometheus.Labels) prometheus.ObserverVec {
	v.touch()
	return v.SummaryVec.MustCurryWith(labels)
}
func (v *SummaryVec) With(labels prometheus.Labels) prometheus.Observer {
	v.touch()
	return v.SummaryVec.With(labels)
}
func (v *SummaryVec) WithLabelValues(lvs ...string) prometheus.Observer {
	v.touch()
	return v.SummaryVec.WithLabelValues(lvs...)
}
