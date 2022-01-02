package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

//定义哈希函数,将[]byte映射成hash值
type Hash func(data []byte) uint32

//一致性hash数据模型
type Map struct {
	//hash函数
	hash Hash
	//虚拟节点倍数
	replicas int
	//hash环
	keys []int
	//虚拟节点和真实节点的映射表(key是虚拟节点的hash值,value是真实节点名称)
	hashMap map[int]string
}

//实例化
func New(replicas int, hash Hash) *Map {
	m := &Map{
		hash:     hash,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	//当hash函数未空时,使用默认hash算法
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

//添加一个或多个节点
func (m *Map) Add(keys ...string) {
	//遍历每个节点
	for _, key := range keys {
		//循环虚拟节点倍数
		for i := 0; i < m.replicas; i++ {
			//算出加上编号后的hash值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			//加入hash环
			m.keys = append(m.keys, hash)
			//增加hash值到真实节点位置的映射
			m.hashMap[hash] = key
		}
	}
	//将hash环排序
	sort.Ints(m.keys)
}

//将key哈希到某节点
func (m *Map) Get(key string) (keyTarget string, ok bool) {
	if len(m.keys) == 0 {
		//当没有节点时
		return
	}
	//算出该key的hash值
	hash := int(m.hash([]byte(key)))
	//二分查找目标节点,返回keys中的下标
	index := sort.Search(len(m.keys), func(i int) bool {
		//当找到节点比该key的hash值大的时候就返回true
		return m.keys[i] >= hash
	})
	//返回该下标对应的hash值对应的真实ip(求下标对应hash值时应考虑index == len时,应取keys[0],所以使用取余来求)
	return m.hashMap[m.keys[index%len(m.keys)]], true
}
