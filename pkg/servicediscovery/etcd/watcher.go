package etcd

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ahfuzhang/cheaprpc/pkg/util/debugs"
	stringsutil "github.com/ahfuzhang/cheaprpc/pkg/util/strings"
)

// func (c *Client) getWatcherByPath(path string){
// 	c.watchers[path]
// }

// RPCClientInfo 为client提供的寻址信息
type RPCClientInfo struct {
	Addr string
	Tags map[string]string
	// todo: 为客户端增加更丰富的统计信息
}

// Watcher etcd watcher 的实现
type Watcher struct {
	Namespace string
	EnvName   string
	App       string
	Server    string
	Service   string
	ch        clientv3.WatchChan
	lock      sync.Mutex
	clients   map[string]*RPCClientInfo
}

// GetPath etcd中的注册路径
func (w *Watcher) GetPath() string {
	return fmt.Sprintf(etcdWatchPath, w.Namespace, w.EnvName, w.App, w.Server, w.Service)
}

func parseValue(v []byte) map[string]string {
	s := stringsutil.NoAllocString(v)
	arr := strings.Split(s, "\n")
	out := make(map[string]string, len(arr))
	for _, row := range arr {
		equal := strings.IndexByte(row, '=')
		if equal == -1 {
			continue
		}
		key := strings.Trim(row[:equal], " ")
		value := strings.Trim(row[equal+1:], " ")
		out[key] = value
	}
	return out
}

const minIPPortLen = 9 // 1.2.3.4:5

// OnChange 某个service的成员发生变化时，回调此函数
func (w *Watcher) OnChange(puts map[string]interface{}, deleted map[string]interface{}) {
	fullPath := w.GetPath()
	for k, v := range puts {
		b, ok := v.([]byte)
		if !ok {
			log.Printf("value not a []byte:%+v\n", v)
			continue
		}
		if len(k) < len(fullPath)+minIPPortLen {
			log.Printf("key format error:%+v\n", k)
			continue
		}
		addr := k[len(fullPath):]
		m := parseValue(b)
		w.lock.Lock()
		w.clients[addr] = &RPCClientInfo{
			Addr: addr,
			Tags: m,
		}
		w.lock.Unlock()
	}
	//
	for k := range deleted {
		if len(k) < len(fullPath)+minIPPortLen {
			log.Printf("key format error:%+v\n", k)
			continue
		}
		addr := k[len(fullPath):]
		w.lock.Lock()
		if _, ok := w.clients[addr]; !ok {
			log.Printf("Addr not exists:%+v\n", k)
			w.lock.Unlock()
			continue
		}
		delete(w.clients, addr)
		w.lock.Unlock()
	}
}

// SelectAll 查询所有节点
func (w *Watcher) SelectAll(c *Client) (map[string]*RPCClientInfo, error) {
	w.lock.Lock()
	out := make(map[string]*RPCClientInfo, len(w.clients))
	if len(w.clients) > 0 {
		for addr, client := range w.clients {
			out[addr] = client
		}
		w.lock.Unlock()
		return out, nil
	}
	w.lock.Unlock()
	// 查询服务器端
	path := w.GetPath()
	values, err := c.getByPrefix(path)
	if err != nil {
		return nil, debugs.WarpError(err, "watcher select all error")
	}
	w.lock.Lock()
	for addr, value := range values {
		if inst, ok := w.clients[addr]; ok {
			out[addr] = inst
			continue
		}
		m := parseValue(stringsutil.NoAllocBytes(value))
		inst := &RPCClientInfo{
			Addr: addr,
			Tags: m,
		}
		w.clients[addr] = inst
		out[addr] = inst
	}
	w.lock.Unlock()
	return out, nil
}

// SelectOne 选择一个客户端
func (w *Watcher) SelectOne(c *Client, exclude map[string]struct{}) (*RPCClientInfo, error) {
	allInst, err := w.SelectAll(c)
	if err != nil {
		return nil, debugs.WarpError(err, "SelectOne error")
	}
	list := make([]string, 0, len(allInst))
	for addr := range allInst {
		if _, ok := exclude[addr]; ok {
			continue
		}
		list = append(list, addr)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("[%s]not have any client", debugs.SourceCodeLoc(1))
	}
	idx := rand.Intn(len(list))
	addr := list[idx]
	return allInst[addr], nil
}
