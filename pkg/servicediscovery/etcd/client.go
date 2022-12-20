// Package etcd 把ETCD封装为服务注册和服务发现的组件
package etcd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"text/template"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ahfuzhang/cheaprpc/pkg/config"
	"github.com/ahfuzhang/cheaprpc/pkg/servicediscovery"
	"github.com/ahfuzhang/cheaprpc/pkg/util/debugs"
	stringsutil "github.com/ahfuzhang/cheaprpc/pkg/util/strings"
)

var (
	//nolint:lll
	etcdPath  = template.Must(template.New("path").Parse(`/cheaprpc/{{.Global.Namespace}}/{{.Global.EnvName}}/{{.Server.App}}/{{.Server.Server}}/{{.Service}}/{{.Addr}}`))
	etcdValue = template.Must(template.New("path").Parse(`namespace={{.Global.Namespace}}
env_name={{.Global.EnvName}}
container_name={{.Global.ContainerName}}
ip={{.Global.LocalIP}}
app={{.Server.App}}
server={{.Server.Server}}
service={{.Service}}
Addr={{.Addr}}
`))
	//nolint:lll
	etcdPathPrefix = template.Must(template.New("path").Parse(`/cheaprpc/{{.Global.Namespace}}/{{.Global.EnvName}}/{{.Server.App}}/{{.Server.Server}}/{{.Service}}/`))
)

const (
	etcdWatchPath = "/cheaprpc/%s/%s/%s/%s/%s/" // namespace, env_name, app, server, service
)

const (
	defaultLeaseTTL = 60
	maxAddrCount    = 1000
)

const defaultBufferSize = 1024

// Client 封装ETCD客户端.
type Client struct {
	cli         *clientv3.Client
	leaseID     clientv3.LeaseID
	lease       clientv3.Lease
	ctx         context.Context //nolint:containedctx
	wg          sync.WaitGroup
	watchers    map[string]*Watcher // path -> watcher 对象
	watcherLock sync.Mutex
}

// NewClient 构造ETCD客户端对象
func NewClient(ctx context.Context, endpoints []string, timeout time.Duration) (*Client, error) {
	c := &Client{
		ctx:      ctx,
		watchers: make(map[string]*Watcher),
	}
	var err error
	c.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, debugs.WarpError(err, "clientv3.New error")
	}
	c.lease = clientv3.NewLease(c.cli)
	var rsp *clientv3.LeaseGrantResponse
	rsp, err = c.lease.Grant(ctx, defaultLeaseTTL)
	if err != nil {
		c.cli.Close()
		return nil, debugs.WarpError(err, "lease.Grant error")
	}
	c.leaseID = rsp.ID
	c.wg.Add(1)
	go c.keepalive()
	return c, nil
}

// keepalive 通过独立的协程，让客户端与ETCD服务器保持心跳
func (c *Client) keepalive() {
	defer c.wg.Done()
	channel, _ := c.lease.KeepAlive(c.ctx, c.leaseID)
	for {
		select {
		case ka := <-channel:
			log.Printf("ttl:%+v,  %+v\n", ka.TTL, ka)
		case <-c.ctx.Done():
			return
		}
	}
}

// Close 释放客户端对象的资源
func (c *Client) Close() {
	c.lease.Close()
	// c.watcherLock.Lock()
	// for _, ch := range c.watchers{
	// 	close(ch)
	// }
	// c.watcherLock.Unlock()
	// todo: 释放 watcher 的资源
	c.wg.Wait() // wait all watch corountine to end
	c.cli.Close()
}

// Register 注册服务
func (c *Client) Register(cfg *config.Configs, service string, addr string) error {
	type param struct {
		*config.Configs
		Service string
		Addr    string
	}
	p := param{Configs: cfg}
	p.Service = service
	p.Addr = addr
	path := bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	if err := etcdPath.Execute(path, &p); err != nil {
		return debugs.WarpError(err, "make etcd path error")
	}
	value := bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	if err := etcdValue.Execute(value, &p); err != nil {
		return debugs.WarpError(err, "make etcd value error")
	}
	_, err := c.cli.Put(c.ctx, path.String(), value.String(), clientv3.WithLease(c.leaseID))
	if err != nil {
		//nolint:wrapcheck
		return debugs.WarpError(err, "etcd put error")
	}
	return nil
}

func (c *Client) getPathOfService(cfg *config.Configs, service string) (string, error) {
	type param struct {
		*config.Configs
		Service string
	}
	p := param{Configs: cfg}
	p.Service = service
	path := bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	if err := etcdPathPrefix.Execute(path, &p); err != nil {
		return "", debugs.WarpError(err, "make etcd path error")
	}
	return path.String(), nil
}

// getByPrefix 根据前缀获取所有内容
func (c *Client) getByPrefix(prefix string) (map[string]string, error) {
	rsp, err := c.cli.Get(c.ctx, prefix, clientv3.WithPrefix(), clientv3.WithLimit(maxAddrCount))
	if err != nil {
		return nil, debugs.WarpError(err, "etcd get error")
	}
	out := make(map[string]string, len(rsp.Kvs))
	for _, kv := range rsp.Kvs {
		k := stringsutil.NoAllocString(kv.Key)
		addr := k[len(prefix):]
		out[addr] = string(kv.Value)
	}
	return out, nil
}

func (c *Client) getWatcherByPath(path string) *Watcher {
	c.watcherLock.Lock()
	w, ok := c.watchers[path]
	c.watcherLock.Unlock()
	if !ok {
		return nil
	}
	return w
}

// SelectAll 查询服务的所有节点
func (c *Client) SelectAll(cfg *config.Configs, service string) (map[string]*RPCClientInfo, error) {
	path, err := c.getPathOfService(cfg, service)
	if err != nil {
		return nil, err
	}
	watcher := c.getWatcherByPath(path)
	if watcher == nil {
		watcher, err = c.AddWatcher(cfg, service)
		if err != nil {
			return nil, debugs.WarpError(err, "AddWatcher error")
		}
	}
	return watcher.SelectAll(c)
}

// SelectOne 查询服务的所有节点
func (c *Client) SelectOne(cfg *config.Configs, service string, exclude map[string]struct{}) (*RPCClientInfo, error) {
	path, err := c.getPathOfService(cfg, service)
	if err != nil {
		return nil, err
	}
	watcher := c.getWatcherByPath(path)
	if watcher == nil {
		watcher, err = c.AddWatcher(cfg, service)
		if err != nil {
			return nil, debugs.WarpError(err, "AddWatcher error")
		}
	}
	return watcher.SelectOne(c, exclude)
}

// Select 查询服务的节点
func (c *Client) Select(cfg *config.Configs, service string, excludes ...string) (*RPCClientInfo, error) {
	excludeSet := make(map[string]struct{}, len(excludes))
	for _, item := range excludes {
		excludeSet[item] = struct{}{}
	}
	return c.SelectOne(cfg, service, excludeSet)
}

// NewWatcher 构造一个新的watcher对象
func (c *Client) NewWatcher(namespace string, envName string, app string, server string,
	service string) (servicediscovery.Watcher, error) {
	w := &Watcher{
		Namespace: namespace,
		EnvName:   envName,
		App:       app,
		Server:    server,
		Service:   service,
		ch:        nil,
		lock:      sync.Mutex{},
		clients:   map[string]*RPCClientInfo{},
	}
	err := c.Watch(w)
	return w, err
}

// NewWatcher 构造一个新的watcher对象
func (c *Client) AddWatcher(cfg *config.Configs,
	service string) (*Watcher, error) {
	w := &Watcher{
		Namespace: cfg.Global.Namespace,
		EnvName:   cfg.Global.EnvName,
		App:       cfg.Server.App,
		Server:    cfg.Server.Server,
		Service:   service,
		ch:        nil,
		lock:      sync.Mutex{},
		clients:   map[string]*RPCClientInfo{},
	}
	err := c.Watch(w)
	return w, err
}

// Watch 观测某个路径的变化
func (c *Client) Watch(w servicediscovery.Watcher) error {
	watcher, ok := w.(*Watcher)
	if !ok {
		return fmt.Errorf("[%s]not a etcd watcher", debugs.SourceCodeLoc(1))
	}
	path := watcher.GetPath()
	ch := c.cli.Watch(c.ctx, path, clientv3.WithPrefix() /*, clientv3.WithLimit(maxAddrCount)*/)
	c.watcherLock.Lock()
	watcher.ch = ch
	c.watchers[path] = watcher
	c.watcherLock.Unlock()
	c.wg.Add(1)
	go c.watchCallBack(watcher)
	return nil
}

// watchCallBack watch的路径发生变化时，在独立的协程中监听此事件
func (c *Client) watchCallBack(watcher *Watcher) {
	defer c.wg.Done()
	for {
		select {
		case rsp := <-watcher.ch:
			if rsp.Err() != nil {
				log.Printf("etcd watch error, err=%+v", rsp.Err())
				continue
			}
			put := make(map[string]interface{})
			deleted := make(map[string]interface{})
			for _, event := range rsp.Events {
				switch event.Type {
				case mvccpb.PUT:
					put[string(event.Kv.Key)] = event.Kv.Value
				case mvccpb.DELETE:
					deleted[string(event.Kv.Key)] = event.Kv.Value
				}
			}
			watcher.OnChange(put, deleted)
		case <-c.ctx.Done():
			return
		}
	}
}
