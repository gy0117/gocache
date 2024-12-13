package cache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/gy0117/gocache/consistenthash"
	"github.com/gy0117/gocache/pb"
	"github.com/gy0117/gocache/peers"
)

const CACHE_BASE_PATH = "/_marscache/"
const REPLICS_PEERS = 100

// 分布式缓存，实现节点间通信
type HttpPool struct {
	hostPort string
	basepath string

	mutex       sync.Mutex
	peersMap    *consistenthash.Map    // 一致性哈希
	httpGetters map[string]*httpGetter // 一个节点对应一个httpGetter
}

func NewHttpPool(hostport string) *HttpPool {
	return &HttpPool{
		hostPort: hostport,
		basepath: CACHE_BASE_PATH,
	}
}

// 添加节点
func (hp *HttpPool) Set(peers ...string) {
	hp.mutex.Lock()
	defer hp.mutex.Unlock()

	hp.peersMap = consistenthash.New(REPLICS_PEERS, nil)
	hp.peersMap.Add(peers...)

	hp.httpGetters = make(map[string]*httpGetter)
	for _, peer := range peers {
		hp.httpGetters[peer] = &httpGetter{
			baseUrl: peer + hp.basepath,
		}
	}
}

// 实现PeerPicker接口，根据key，找到对应的节点，然后根据节点，找到对应的PeerGetter
func (hp *HttpPool) PickPeer(key string) (peers.PeerGetter, bool) {
	hp.mutex.Lock()
	defer hp.mutex.Unlock()

	peer := hp.peersMap.Get(key)
	if peer != "" && peer != hp.hostPort {
		return hp.httpGetters[peer], true
	}
	return nil, false
}

//  1. 解析url，拿到groupname和key
//     判断path是否是以/_geecache/为前缀的。 否，则panic
//  2. 根据group和key，获取到对应的value，然后写到writer中
//
// 例如：http://127.0.0.1/_marscache/users/zhangsan，groupname是users，key是zhangsan，即获取users group下的key为zhangsan对应的value
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basepath) {
		panic("url path is not '/_geecache/' as prefix.")
	}

	path := r.URL.Path
	// /_marscache/scores/Tom
	log.Printf("HttpPool.ServeHTTP | path:%v\n", path[len(CACHE_BASE_PATH):])
	parts := strings.SplitN(path[len(CACHE_BASE_PATH):], "/", 2)

	groupname := parts[0]
	key := parts[1]

	log.Printf("HttpPool.ServeHTTP | group_name: %v, key: %v\n", groupname, key)

	g := GetGroup(groupname)

	item, err := g.Get(key)
	if err != nil {
		log.Printf("HttpPool.ServeHTTP | g.Get | key: %v, err: %+v\n", key, err)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: item.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)

}

// 客户端实现PeerGetter接口
type httpGetter struct {
	baseUrl string // 例如：http://127.0.0.1/_marscache/
}

// 1. 拼接url，执行请求
// 替换成下面的gRPC通信
// func (hg *httpGetter) Get(group string, key string) ([]byte, error) {
// 	// url.QueryEscape对string进行转义
// 	url := fmt.Sprintf("%v%v/%v", hg.baseUrl, url.QueryEscape(group), url.QueryEscape(key))

// 	log.Printf("httpGetter.Get | url: %v\n", url)

// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("server returned: %v", resp.Status)
// 	}

// 	b, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("reading response body: %v", err)
// 	}

// 	return b, nil

// }

func (hg *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	// url.QueryEscape对string进行转义
	url := fmt.Sprintf("%v%v/%v", hg.baseUrl, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))

	log.Printf("httpGetter.Get | url: %v\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err := proto.Unmarshal(b, out); err != nil {
		return fmt.Errorf("proto.Unmarshal response body: %v", err)
	}

	return nil
}
