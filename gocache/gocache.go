package gocache

import (
	"fmt"
	"log"
	"sync"
)

//Getter用于从数据源获取数据
type Getter interface {
	//有一个Get方法
	Get(key string) ([]byte, error)
}

//定义函数类型GetterFunc,并实现Getter接口中的Get方法(接口型函数)
type GetterFunc func(key string) ([]byte, error)

//实现Get方法
func (g GetterFunc) Get(key string) ([]byte, error) {
	//调用自己
	return g(key)
}

//一个group,用于处理缓存查询/添加
type Group struct {
	//当前group名称
	name string
	//从数据源获取数据Getter
	getter Getter
	//当前group的缓存
	mainCache cache
	//远程节点数据的Picker
	peers PeerPicker
}

var (
	//读写锁
	lock sync.RWMutex
	//所有group的集合
	groups = make(map[string]*Group)
)

//初始化一个Group(有并发问题,需要加锁)
func NewGroup(name string, maxBytes int64, getter Getter) *Group {
	if getter == nil {
		//当没有getter时报错
		panic("nil Getter")
	}
	lock.Lock()
	defer lock.Unlock()
	//实例化
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: maxBytes,
		},
	}
	//加入到group集合
	groups[name] = g
	return g
}

//获取Group(并发访问groups,需要加读锁)
func GetGroup(name string) *Group {
	//读锁
	lock.RLock()
	g := groups[name]
	lock.RUnlock()
	return g
}

//Get方法,获取缓存数据
func (g *Group) Get(key string) (ByteView, error) {
	//判空
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	//先从本地获取
	if v, ok := g.mainCache.get(key); ok {
		//缓存命中
		log.Println("gocache hit")
		return v, nil
	}

	//再从远程获取
	return g.load(key)
}

//从远程获取
func (g *Group) load(key string) (value ByteView, err error) {
	//先从远程节点获取
	if g.peers != nil {
		if peerGetter, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peerGetter, key); err == nil {
				//当peerGetter获取数据无异常时
				return value, nil
			}
			//远程节点获取数据错误时
			log.Println("[GoCache] Failed to get from peer", err)
		}
	}
	//本地数据源获取
	return g.getLocally(key)
}

//从本地获取
func (g *Group) getLocally(key string) (ByteView, error) {
	//使用Get方法来从数据源中获取数据
	v, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	//获取value
	value := ByteView{
		b: cloneByte(v),
	}
	//将缓存添加到本地
	g.addCacheLocal(key, value)
	return value, nil
}

//将缓存添加到本地
func (g *Group) addCacheLocal(key string, value ByteView) {
	g.mainCache.put(key, value)
}

//注册Picker
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

//从远程节点获取数据
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes}, nil
}
