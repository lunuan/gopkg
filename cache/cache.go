package cache

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
)

var cache *bigcache.BigCache

func Init() {
	config := bigcache.Config{
		// Shards默认值为1024，适用于50GB的最大缓存大小。但是如果我们的缓存空间不大，可以考虑减少Shards数量，增加每个Shard内缓存的大小，并使用更少的CPU时间来计算哈希值
		Shards: 16,
		// LifeWindow表示缓存项生存期，在这段时间之后如果还没有被访问，就会被从缓存中删除。这里设置的是10分钟，可以根据需求进行调整。要考虑到缓存时间过长可能会导致缓存命中率下降，过短可能会导致缓存击穿
		LifeWindow: 10 * time.Minute,
		// CleanWindow表示缓存清理时间间隔，每隔这么长时间就会检查一遍缓存，清理掉过期的缓存项。这里设置的是1分钟，可以根据实际情况进行调整
		CleanWindow: 1 * time.Minute,
		// MaxEntriesInWindow表示在LifeWindow时间内最多可以缓存的项数。这里设置的是每秒最多可以缓存1000个项，即10分钟内最多可以缓存600000个项。
		MaxEntriesInWindow: 1000 * 10 * 60,
		// MaxEntrySize表示缓存项的最大长度，这里设置的是500字节。可以根据实际情况进行调整。
		MaxEntrySize: 500,
		// Verbose表示是否在日志中记录详细的日志信息。这里设置为false，表示不记录详细日志信息。
		Verbose: false,
		// HardMaxCacheSize表示缓存的最大大小，通常是未使用的物理内存的一小部分。这个值的设置需要考虑到我们系统中可用的内存大小。如果设置得太大，可能导致内存不足。
		HardMaxCacheSize: 8192,
		// OnRemove表示缓存项被删除时的回调函数，这里设置为nil，表示不执行任何操作。
		OnRemove: nil,
		// OnRemoveWithReason表示缓存项被删除时的回调函数，这里设置为nil，表示不执行任何操作。
		OnRemoveWithReason: nil,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if c, err := bigcache.New(ctx, config); err != nil {
		panic(err)
	} else {
		cache = c
	}
}

func Set(key string, value []byte) error {
	return cache.Set(key, value)
}

func Get(key string) ([]byte, error) {
	return cache.Get(key)
}

func Delete(key string) error {
	return cache.Delete(key)
}

func Reset() error {
	return cache.Reset()
}
