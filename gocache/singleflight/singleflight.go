package singleflight

import "sync"

//正在进行或者已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

//管理不同key的call
type Group struct {
	lock sync.Mutex
	m    map[string]*call
}

//针对相同的key,无论Do被调用多少次,函数fn只会被调用一次,等fn调用结束之后返回返回值或者错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	//并发下保护成员变量m,需要加锁
	g.lock.Lock()
	if g.m == nil {
		//map延迟初始化
		g.m = make(map[string]*call)
	}
	//从m中获取目前正在请求的key对应的call请求
	if c, ok := g.m[key]; ok {
		//如果有该key,那么代表已经有一个查询该key的请求过去了,现在只需要阻塞等待
		g.lock.Unlock()
		//在第一个查询该key的call上等待
		c.wg.Wait()
		return c.val, c.err
	}
	//如果没有该key对应的call请求
	c := new(call)
	//设waitGroup计数器为1
	c.wg.Add(1)
	//加入到key->call的映射
	g.m[key] = c
	//解锁
	g.lock.Unlock()
	//调用函数并将值设进call中
	c.val, c.err = fn()
	//该函数调用完成之后,将waitGroup减1,即使相同的key的查询请求不再被阻塞
	c.wg.Done()
	//再加锁删去该key->call的映射
	g.lock.Lock()
	delete(g.m, key)
	g.lock.Unlock()
	return c.val, c.err
}
