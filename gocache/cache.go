package gocache

import (
	"gocache/lru"
	"sync"
)

const (
	maxcachesize = 1 << 10
)

type cache struct {
	mu  *sync.RWMutex
	lru *lru.LRU
}

func (c *cache) Add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.NewLRU(maxcachesize, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) Get(key string) (value ByteView, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.lru == nil {
		return
	}
	var v interface{}
	if v, ok = c.lru.Get(key); !ok {
		return
	}
	return v.(ByteView), true
}
