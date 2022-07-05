package gocache

import (
	"fmt"
	"gocache/lru"
	"sync"
)

type Group struct {
	mainCache cache
	name      string
	getter    Getter
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

var (
	mu     *sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, capacity int, getter Getter) *Group {
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		mainCache: cache{
			mu:  &sync.RWMutex{},
			lru: lru.NewLRU(capacity, nil),
		},
		name:   name,
		getter: getter,
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.Unlock()
	return groups[name]
}

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key required")
	}
	v, ok := g.mainCache.Get(key)
	if !ok {
		return g.load(key)
	}
	return v, nil
}

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	var (
		bytes []byte
		err   error
	)
	if bytes, err = g.getter.Get(key); err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}
