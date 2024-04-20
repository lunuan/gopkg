package pushservice

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	taskPushTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pushservice_task_push_total",
		Help: "pushservice task push total",
	}, []string{"collector"})
	serviceSelfInspectionTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pushservice_self_inspection_total",
		Help: "pushservice self inspection total",
	})
)

type PushServiceCollector struct {
	service *PushService
}

func NewPushServiceCollector(s *PushService) *PushServiceCollector {
	c := &PushServiceCollector{
		service: s,
	}
	return c
}

func (c *PushServiceCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *PushServiceCollector) Collect(ch chan<- prometheus.Metric) {
	logger.Debug("collect metrics")

	// registered collector count in push service
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("pushservice_collector_count", "registered collector count in push service", nil, nil),
		prometheus.GaugeValue,
		float64(len(c.service.collectors)),
	)

	// registered task count in push service
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("pushservice_task_count", "registered task count in push service", nil, nil),
		prometheus.GaugeValue,
		float64(len(c.service.tasks)),
	)

	// collect push service status
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("pushservice_status", "pushservice status", nil, nil),
		prometheus.GaugeValue,
		c.collectServiceStatus(),
	)

	// collect task status
	taskStatusDesc := prometheus.NewDesc("pushservice_task_status", "task status", []string{"collector"}, nil)
	for idx, task := range c.service.tasks {
		ch <- prometheus.MustNewConstMetric(
			taskStatusDesc,
			prometheus.GaugeValue,
			c.collectTaskStatus(task),
			idx,
		)
	}

	// collect task push total
	taskPushTotal.Collect(ch)

	// collect self inspection total
	serviceSelfInspectionTotal.Collect(ch)
}

func (c *PushServiceCollector) collectServiceStatus() float64 {
	switch c.service.status {
	case "init":
		return 0.0
	case "running":
		return 1.0
	case "stopped":
		return 2.0
	default:
		return -1.0
	}
}

func (c *PushServiceCollector) collectTaskStatus(task *PushTask) float64 {
	if task.running {
		return 1.0
	}
	return 0.0
}
