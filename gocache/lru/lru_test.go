package lru

import (
	"fmt"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

//测试查询数据
func TestCache_Get(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Put("key1", String("1111"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1111" {
		t.Fatal("cache git key1 = 1111 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatal("cache miss key2 failed")
	}
}

//测试删除最后一个
func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Put(k1, String(v1))
	lru.Print()
	lru.Put(k2, String(v2))
	lru.Print()
	lru.Put(k3, String(v3))
	lru.Print()
	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

//测试回调方法
func TestOnEvicted(t *testing.T) {
	callback := func(key string, value Value) {
		fmt.Println("delete " + key + "->" + string(value.(String)))
	}
	lru := New(int64(10), callback)
	lru.Put("k1", String("123456"))
	lru.Print()
	lru.Put("k2", String("k2"))
	lru.Print()
	lru.Put("k3", String("k3"))
	lru.Print()
	lru.Put("k4", String("k4"))
	lru.Print()
}
