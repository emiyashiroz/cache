package main

import (
	"fmt"
	"gocache"
	"log"
	"net/http"
)

var (
	db = map[string]string{"zw": "1231ed", "qiuqiu": "sdas", "weui": "12e9i8dhsh"}
)

type localGetter struct {
}

func (lg *localGetter) Get(key string) ([]byte, error) {
	v, ok := db[key]
	if !ok {
		return nil, fmt.Errorf("key invalid")
	}
	return []byte(v), nil
}

func main() {
	gocache.NewGroup("local", 2, &localGetter{})
	addr := "localhost:9999"
	peers := gocache.NewHttpTool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
