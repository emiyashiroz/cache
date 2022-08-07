package gocache

import (
	"fmt"
	pb "gocache/cachepb"
	"gocache/lru"
	"gocache/singleflight"
	"log"
	"sync"
)

type Group struct {
	mainCache cache
	name      string
	getter    Getter
	peers     PeerPicker
	loader    *singleflight.Group // use singleflight.Group to make sure that each key is only fetched once

}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

var (
	mu     = &sync.RWMutex{}
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
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
		loader: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key required")
	}
	if v, ok := g.mainCache.Get(key); ok {
		log.Println("[GoCache] hit")
		return v, nil
	}
	return g.load(key)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}
func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GoCache] failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
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

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
