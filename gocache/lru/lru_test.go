package lru

import (
	"fmt"
	"testing"
)

func TestLRU(t *testing.T) {
	lru := NewLRU(2, func(key string, value Value) {
		fmt.Println(key)
	})
	lru.Add("zw", string("223"))
	lru.Add("key", string("dasd"))
	lru.Add("sd", string("sdsd"))
	lru.Get("key")
	lru.Add("dasdsd", string("sdsd"))
	_, ok1 := lru.Get("sd")
	v2, _ := lru.Get("key")
	v3, _ := lru.Get("dasdsd")
	if ok1 == false && string(v2.(string)) == "dasd" && string(v3.(string)) == "sdsd" {
		fmt.Println("test success")
	}
}
