package pushservice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lunuan/gopkg/log"
	"github.com/lunuan/gopkg/prometheus"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/push"
)

var logger *zap.SugaredLogger

func init() {
	logger = log.NewSugaredLogger(&log.Config{Level: "debug"}).Named("pushservice")
}

type PushTask struct {
	url       string
	interval  time.Duration
	running   bool
	cancel    *context.CancelFunc
	collector prometheus.Collector
	mux       *sync.Mutex
}

func newPushTask(url string, interval time.Duration, collector prometheus.Collector) *PushTask {
	logger.Debugw("new push task", "url", url, "interval", interval, "collector", collector)
	return &PushTask{
		url:       url,
		interval:  interval,
		running:   false,
		collector: collector,
		mux:       &sync.Mutex{},
	}
}

func (t *PushTask) run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorw("recovered", "collector", t.collector.Idx(), "panic", r)
		}
		t.running = false
		t.mux.Unlock()
	}()
	t.mux.Lock()
	t.running = true
	logger.Infow("task started", "collector", t.collector.Idx())
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = &cancel
	pusher := push.New(t.url, fmt.Sprintf("pushservice_%s", t.collector.Idx()))
	pusher.Collector(t.collector)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := pusher.Push()
			if err != nil {
				logger.Error(err.Error())
			}
			taskPushTotal.WithLabelValues(t.collector.Idx()).Inc()
		}
		logger.Debugw("push request success", "collector", t.collector.Idx())
		time.Sleep(t.interval)
	}
}

var taskStopedCause = map[int]string{
	1: "collector expired",
	2: "collector unregistered",
	3: "pushservice stoped",
	4: "task orphaned",
}

/*
stop stops the push task, and delete metrics from pushgateway if the task is stopped by expired collector or unregister collector

	cause code:
	1: "collector expired",
	2: "collector unregistered",
	3: "pushservice stoped",
	4: "task orphaned",
*/
func (t *PushTask) stop(cause int) {
	if t.cancel != nil {
		(*t.cancel)()
	}
	t.running = false

	logger.Infow("task stoped", "collector", t.collector.Idx(), "cause", taskStopedCause[cause])

	// delete metrics from pushgateway if the task is stopped by expired collector or unregister collector
	if cause == 1 || cause == 2 {
		pusher := push.New(t.url, "pushservice")
		pusher.Collector(t.collector)
		err := pusher.Delete()
		if err != nil {
			logger.Errorw("failed to delete metrics in pushgateway", "collector", t.collector.Idx(), "error", err)
			return
		}
		logger.Infow("deleted from pushgateway", "collector", t.collector.Idx())
	}
}

type PushService struct {
	url        string                          // url is the pushgateway url
	interval   time.Duration                   // interval is the interval of pushing metrics
	expired    time.Duration                   // expired is the duration of expiration
	status     string                          // status is the status of the push service, used to determine the status of push services during health checks
	collectors map[string]prometheus.Collector // collectors is the map of collectors
	tasks      map[string]*PushTask            // tasks is the map of push tasks
	mutex      *sync.Mutex
}

func NewPushService(url string, interval time.Duration, expired time.Duration) *PushService {
	ps := &PushService{
		url:        url,
		interval:   interval,
		expired:    expired,
		status:     "init",
		collectors: make(map[string]prometheus.Collector),
		tasks:      make(map[string]*PushTask),
		mutex:      &sync.Mutex{},
	}
	go ps.start()
	return ps
}

func (s *PushService) start() {
	s.status = "running"
	defer func() {
		if r := recover(); r != nil {
			logger.Errorw("recovered from panic", "panic", r)
		}
		s.status = "stopped"
	}()
	logger.Infow("push service started", "url", s.url, "interval", s.interval, "expired", s.expired)
	for {
		// check push service status for each collector every interval duration
		logger.Debugw("check push task status for each collector", "expired", s.expired)
		s.mutex.Lock()
		for _, c := range s.collectors {
			task := s.tasks[c.Idx()]

			if c.Expired(s.expired) && task.running {
				task.stop(1)
			} else if !c.Expired(s.expired) && !task.running {
				logger.Infof("task %s is unexpired but not running, restart", c.Idx())
				go task.run()
			}
		}

		for id, task := range s.tasks {
			if _, ok := s.collectors[id]; !ok {
				logger.Warnw("orphan task found", "task", id)
				task.stop(4)
				delete(s.tasks, id)
			}
		}

		s.mutex.Unlock()
		serviceSelfInspectionTotal.Inc()
		time.Sleep(s.interval)
	}
}

// Register registers the collector to the push service
func (s *PushService) Register(c prometheus.Collector) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.collectors[c.Idx()] = c
	task := newPushTask(s.url, s.interval, c)
	s.tasks[c.Idx()] = task
	go task.run()
	logger.Infow("collector registered", "idx", c.Idx())
}

// Unregister unregisters the collector from the push service
func (s *PushService) Unregister(c prometheus.Collector) {
	logger.Infow("unregister collector", "collector", c.Idx())
	defer func() {
		if r := recover(); r != nil {
			logger.Errorw("recovered from panic", "panic", r)
			return
		}
		logger.Infof("unregister collector %d complete", c.Idx())
	}()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if c, ok := s.collectors[c.Idx()]; ok {
		delete(s.collectors, c.Idx())
	}

	if task, ok := s.tasks[c.Idx()]; ok {
		task.stop(2)
		delete(s.tasks, c.Idx())
	}
}

func (s *PushService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	wg := &sync.WaitGroup{}
	for _, task := range s.tasks {
		wg.Add(1)
		task.stop(3)
		wg.Done()
	}
	wg.Wait()
	logger.Infof("push service stopped")
}
