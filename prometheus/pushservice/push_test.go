package pushservice

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/lunuan/gopkg/prometheus"

	prometheus_client "github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
)

func TestPush(t *testing.T) {
	counterVec := prometheus.NewCounterVec(prometheus_client.CounterOpts{
		Name: "zzzz_counter_vec",
		Help: "This is my counter vec metric",
	}, []string{"label1", "label2"})
	// prometheus.MustRegister(counterVec)

	go func() {
		for {
			counterVec.WithLabelValues("value1", "value2").Add(10)
			counterVec.WithLabelValues("value1", "value2").Inc()
			time.Sleep(1 * time.Second)
		}
	}()

	pusher := push.New("http://127.0.0.1:9091", "push_test")
	pusher.Grouping("instance", "localhost")
	pusher.Collector(counterVec)

	for i := 0; i < 10; i++ {
		err := pusher.Push()
		if err != nil {
			fmt.Println("failed to push metrics:", err)
			return
		}
		fmt.Println("success to push metrics")
		time.Sleep(time.Second)
	}

	// pusher.Delete()
}

func TestPushService(t *testing.T) {
	// push metrics to pushgateway every 1 second
	// if the collector is not updated for 10 seconds, stop pushing
	pushService := NewPushService("http://127.0.0.1:9091", time.Second*15, time.Hour*4)

	counter := prometheus.NewCounter(prometheus_client.CounterOpts{
		Name: "zzzz_counter_sample",
		Help: "This is my counter metric",
	})
	pushService.Register(counter)

	counterVec := prometheus.NewCounterVec(prometheus_client.CounterOpts{
		Name: "zzzz_counter_vec",
		Help: "This is my counter vec metric",
	}, []string{"label1", "label2"})
	pushService.Register(counterVec)

	counter.Add(100)
	counterVec.WithLabelValues("value1", "value2").Add(200)

	c := NewPushServiceCollector(pushService)
	prometheus_client.MustRegister(c)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)

}
