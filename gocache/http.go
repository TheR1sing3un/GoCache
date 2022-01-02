package gocache

import (
	"fmt"
	"github.com/TheR1sing3un/gocache/gocache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	//默认路径前缀
	defaultBasePath = "/cache"
	//默认虚拟节点倍数
	defaultReplicas = 50
)

//定义Http连接实体(集成该节点的服务端和其他节点的本地客户端)
type HttpPool struct {
	//自己的地址(主机名/IP和端口)
	self string
	//节点间通讯地址前缀
	basePath string
	//互斥锁
	lock sync.Mutex
	//一致性哈希模型
	peers *consistenthash.Map
	//远程节点的本地客户端集合(url->httpGetter实例)
	httpGetters map[string]*httpGetter
}

//实例化方法
func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

//自定义的日志记录方法
func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

//Serve方法
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//当路径前缀不符合时报错
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HttpPool serving unexpected path : " + r.URL.Path)
	}
	//记录请求的类型和路径
	p.Log("%s %s", r.Method, r.URL.Path)
	//请求路径格式<basePath>/<groupName>/<key>
	//将path从baseUrl后开始切割成两份,分别是groupName和key
	parts := strings.SplitN(r.URL.Path[len(p.basePath)+1:], "/", 2)
	if len(parts) != 2 {
		//返回错误信息
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	//根据group名获取group
	group := GetGroup(groupName)
	if group == nil {
		//当获取到组名为空
		http.Error(w, "group: "+groupName+"no found", http.StatusNotFound)
		return
	}
	//从组中获取数据
	v, err := group.Get(key)
	//异常
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//设置响应头
	w.Header().Set("Content-Type", "application/octet-stream")
	//写回
	w.Write(v.ByteSlice())
}

//定义http客户端,和远程节点的服务端交互
type httpGetter struct {
	baseURL string
}

//实现PeerGetter接口的Get方法,从远程节点获取数据
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	//拼接节点url
	u := fmt.Sprintf("%v/%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	//最后需要关闭resp的返回体
	defer resp.Body.Close()

	//如果返回的不是200
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned : %v", resp.Status)
	}

	//读取返回体
	v, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("reading response body error : %v", err)
	}

	return v, nil
}

//设置远程节点(传入主机名/ip:port)
func (p *HttpPool) Set(peers ...string) {
	//改变成员变量(peers和httpGetters),因此需要上锁
	p.lock.Lock()
	defer p.lock.Unlock()
	//初始化一致性哈希模型
	p.peers = consistenthash.New(defaultReplicas, nil)
	//将远程节点加入到一致性哈希模型中
	p.peers.Add(peers...)
	//实例化httpGetters
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	//循环创建每个节点的httpGetter
	for _, peer := range peers {
		//实例化httpGetter,并加入到集合
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

//根据key来获取目标节点的PeerGetter(也就是这里的httpGetter)
func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	//有并发问题,需要加锁
	p.lock.Lock()
	defer p.lock.Unlock()
	//从一致性哈希模型中获取该key对应的节点
	//当获取到该key映射到的节点地址,并且该地址不为空不为本机
	if peer, ok := p.peers.Get(key); ok && peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
