package lru

type Value interface {
}

type Entry struct {
	key   string
	value Value
}

type Node struct {
	pre   *Node
	next  *Node
	entry Entry
}

type LRU struct {
	head      *Node
	tail      *Node
	capacity  int
	key2node  map[string]*Node
	onEvicted func(key string, value Value)
}

func NewLRU(capacity int, onEvicted func(key string, value Value)) *LRU {
	head := &Node{}
	tail := &Node{}
	head.next = tail
	tail.pre = head
	return &LRU{
		head:      head,
		tail:      tail,
		capacity:  capacity,
		key2node:  map[string]*Node{},
		onEvicted: onEvicted,
	}
}

func (lru *LRU) Get(key string) (interface{}, bool) {
	node, ok := lru.key2node[key]
	if !ok {
		return nil, false
	}
	lru.removeNode(node)
	lru.add2Head(node)
	return node.entry.value, true
}

func (lru *LRU) Add(key string, value Value) {
	node, ok := lru.key2node[key]
	if ok {
		node.entry.value = value
		lru.removeNode(node)
		lru.add2Head(node)
		return
	}
	if lru.Len() == lru.capacity {
		lru.RemoveOldest()
	}
	node = &Node{
		entry: Entry{key, value},
	}
	lru.key2node[key] = node
	lru.add2Head(node)
}

func (lru *LRU) RemoveOldest() {
	if lru.Len() < 1 {
		return
	}
	node := lru.tail.pre
	delete(lru.key2node, node.entry.key)
	lru.removeNode(node)
	if lru.onEvicted != nil {
		lru.onEvicted(node.entry.key, node.entry.value)
	}
}

func (lru *LRU) Len() int {
	return len(lru.key2node)
}

func (lru *LRU) removeNode(node *Node) {
	node.next.pre = node.pre
	node.pre.next = node.next
}

func (lru *LRU) add2Head(node *Node) {
	node.next = lru.head.next
	node.pre = lru.head
	node.pre.next = node
	node.next.pre = node
}
