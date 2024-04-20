package prometheus

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector Collector is a wrapper around prometheus.Collector, which adds the concept of expiration
type Collector interface {
	prometheus.Collector

	Idx() string                         // UID is the unique identifier of the collector
	touch()                              // touch updates the lastAccessedTime of the collector
	Expired(duration time.Duration) bool // Expired checks if the collector is expired if the lastAccessedTime is older than duration
}

// metadata basic metadata of prometheus.Collector
type collectorMetadata struct {
	idx              string    // idx is used to identify the collector
	lastAccessedTime time.Time // updateAt is the time when the collector is updated
}

func (c *collectorMetadata) Idx() string {
	return c.idx
}

func (c *collectorMetadata) touch() {
	c.lastAccessedTime = time.Now()
}

func (c *collectorMetadata) Expired(duration time.Duration) bool {
	return time.Since(c.lastAccessedTime) > duration
}

func NewCollectorMetadata() *collectorMetadata {
	var randint int64 = int64(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100)) * 10
	return &collectorMetadata{
		idx:              strconv.FormatInt(time.Now().UnixNano()+randint, 36),
		lastAccessedTime: time.Now(),
	}
}
