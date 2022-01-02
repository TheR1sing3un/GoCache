package lru

import (
	"container/list"
	"fmt"
)

type Cache struct {
	//缓存队列最大缓存大小
	maxBytes int64
	//缓存队列目前已经使用大小
	usedBytes int64
	//链表
	cacheList *list.List
	//map映射key->Element
	cacheMap map[string]*list.Element
	//删除记录时的回调函数
	OnEvicted func(key string, value Value)
}

//键值对类型
type entry struct {
	key   string
	value Value
}

//缓存的值的类型
type Value interface {
	//返回Value的内存大小
	Len() int
}

//实例化函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		usedBytes: 0,
		cacheList: list.New(),
		cacheMap:  make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//查找
func (c *Cache) Get(key string) (Value, bool) {
	//先从map里查找是否有该值
	if element, ok := c.cacheMap[key]; ok {
		//若有则将该提至最近使用的
		c.cacheList.MoveToFront(element)
		//返回(将element的value强转成自定义的Value类型
		e := element.Value.(*entry)
		return e.value, true
	}
	return nil, false
}

//删除最久未被使用的节点
func (c *Cache) RemoveOldest() {
	//获取最后一个节点
	cur := c.cacheList.Back()
	if cur != nil {
		//从list中删除
		c.cacheList.Remove(cur)
		//从map中删除
		entry := cur.Value.(*entry)
		delete(c.cacheMap, entry.key)
		//更新当前已用内存
		c.usedBytes -= int64(len(entry.key)) + int64(entry.value.Len())
		//运行回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(entry.key, entry.value)
		}
	}
}

//添加缓存
func (c *Cache) Put(key string, value Value) {
	//判断当前是否已经有该key
	if element, ok := c.cacheMap[key]; ok {
		//已经有该节点,则将该节点更新并移至队首(最近访问)
		c.cacheList.MoveToFront(element)
		entry := element.Value.(*entry)
		//更新当前已用内存(加上当前新增的value大小再减去原本的value大小)
		c.usedBytes += int64(value.Len()) - int64(entry.value.Len())
		//更新值
		entry.value = value
	} else {
		//若不存在,则添加至队首,并更新map
		element := c.cacheList.PushFront(&entry{key, value})
		//添加至map
		c.cacheMap[key] = element
		//更新已用内存
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
	//添加之后判断是否已经超过最大内存(当最大内存不为0时而且已用内存超过最大内存时删除最近未使用的节点)
	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.RemoveOldest()
	}
}

//获取当前缓存的数据条数
func (c *Cache) Len() int {
	return c.cacheList.Len()
}

func (c *Cache) Print() {
	for cur := c.cacheList.Front(); cur != nil; cur = cur.Next() {
		kv := cur.Value.(*entry)
		fmt.Printf("%s->%s ", kv.key, kv.value)
	}
	fmt.Println()
}
