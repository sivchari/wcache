package wcache

import (
	"log/slog"
	"runtime"
	"sync"
	"weak"
)

type Cacher[K comparable, V any] interface {
	Get(key K) *V
	Set(key K, value *V)
}

type WeakCache[K comparable, V any] struct {
	mem map[K]weak.Pointer[V]
	mu  sync.RWMutex
	log *slog.Logger
}

var _ Cacher[interface{}, interface{}] = &WeakCache[interface{}, interface{}]{}

func New[K comparable, V any]() *WeakCache[K, V] {
	return &WeakCache[K, V]{
		mem: make(map[K]weak.Pointer[V]),
		log: slog.New(slog.DiscardHandler),
	}
}

func (c *WeakCache[K, V]) WithLogger(log *slog.Logger) *WeakCache[K, V] {
	c.log = log
	return c
}

func (c *WeakCache[K, V]) Get(key K) *V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if ptr, ok := c.mem[key]; ok {
		return ptr.Value()
	}
	return nil
}

func (c *WeakCache[K, V]) Set(key K, value *V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	runtime.AddCleanup(value, c.delete, key)
	c.mem[key] = weak.Make(value)
}

func (c *WeakCache[K, V]) delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.log.Info("deleting key", slog.Any("key", key))
	delete(c.mem, key)
}
