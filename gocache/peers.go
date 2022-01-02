package gocache

type PeerPicker interface {
	//根据key获取相应的peer
	PickPeer(key string) (peer PeerGetter, ok bool)
}

//PeerGetter接口,需要被Peer实现方法
type PeerGetter interface {
	//从group中查找缓存值
	Get(group string, key string) ([]byte, error)
}
