package gocache

import (
	"github.com/TheR1sing3un/gocache/gocache/lru"
	"sync"
)

type cache struct {
	//互斥锁
	lock sync.Mutex
	//lru缓存队列
	lru *lru.Cache
	//最大缓存大小
	cacheBytes int64
}

//缓存put方法
func (c *cache) put(key string, value ByteView) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.lru == nil {
		//懒加载lru
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Put(key, value)
}

//缓存get方法
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.lru == nil {
		//还未初始化(当前肯定没有数据)
		return
	}
	if value, ok := c.lru.Get(key); ok {
		return value.(ByteView), true
	}
	return
}
