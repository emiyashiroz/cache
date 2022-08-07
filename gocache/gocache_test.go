package gocache

import (
	"fmt"
	"testing"
)

var (
	db map[string]string
)

type localGetter struct {
}

func (lg *localGetter) Get(key string) ([]byte, error) {
	v, ok := db[key]
	if !ok {
		return []byte{}, fmt.Errorf("invalid key")
	}
	return []byte(v), nil
}

func TestGocache(t *testing.T) {
	db = make(map[string]string)
	db["1"] = "1"
	db["2"] = "2"
	db["3"] = "3"
	db["4"] = "4"
	g := NewGroup("zw", 2, &localGetter{})
	v1, _ := g.Get("1")
	fmt.Println(v1)
	v2, _ := g.Get("2")
	fmt.Println(v2)
	v3, _ := g.Get("3")
	fmt.Println(v3)
	v4, _ := g.Get("4")
	fmt.Println(v4)

}
