package test

import (
	_ "embed"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/ahfuzhang/cheaprpc/pkg/util/debugs"
	"github.com/ahfuzhang/cheaprpc/pkg/util/strings"
)

//go:embed test_config.yaml
var testConfigStr string

type Configs struct {
	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
	ETCD struct {
		Addrs []string `yaml:"addrs"`
	} `yaml:"etcd"`
}

var (
	cfg  Configs
	once sync.Once
)

func LoadConfig() error {
	var err error
	once.Do(func() {
		err = yaml.Unmarshal(strings.NoAllocBytes(testConfigStr), &cfg)
	})
	if err != nil {
		return debugs.WarpError(err, "yaml decode error")
	}
	return nil
}

func GetConfig() *Configs {
	return &cfg
}
