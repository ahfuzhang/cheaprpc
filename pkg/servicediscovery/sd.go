package servicediscovery

import (
	"github.com/ahfuzhang/cheaprpc/pkg/config"
)

// Watcher 如果要动态监听变化，实现此接口
type Watcher interface {
	GetPath() string
	OnChange(put map[string]interface{}, deleted map[string]interface{})
}

// ServiceDiscovery 实现服务发现的组件
type ServiceDiscovery interface {
	Register(cfg *config.Configs, service string, addr string) error
	Select(cfg *config.Configs, service string) (map[string]string, error)
	Watch(w Watcher) error
}

// type SD struct {
// 	client ServiceDiscovery
// }
//
// func NewServiceDiscovery() (ServiceDiscovery, error) {
// 	return nil, nil
// }
