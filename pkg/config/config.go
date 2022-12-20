package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v3"

	"github.com/ahfuzhang/cheaprpc/pkg/util/debugs"
	stringsutil "github.com/ahfuzhang/cheaprpc/pkg/util/strings"
)

// Global 服务的全局配置.
type Global struct {
	Namespace     string `yaml:"namespace"`
	EnvName       string `yaml:"env_name"`
	ContainerName string `yaml:"container_name"`
	LocalIP       string `yaml:"local_ip"`
}

// Service 服务的配置.
type Service struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
}

// Service 服务的管理端口
type Admin struct {
	IP             string `yaml:"ip"`
	Port           int32  `yaml:"port"`
	ReadTimeoutMs  int64  `yaml:"read_timeout_ms"`
	WriteTimeoutMs int64  `yaml:"write_timeout_ms"`
}

// Server 服务器的启动配置.
type Server struct {
	App     string    `yaml:"app"`
	Server  string    `yaml:"server"`
	Service []Service `yaml:"service"`
	Admin   Admin     `yaml:"admin"`
}

// Zap 日志库zap的配置.
type Zap struct {
	Level          string `yaml:"level"` // debug info warn error dpanic panic fatal
	LogPath        string `yaml:"log_path"`
	Format         string `yaml:"format"` // text json
	MaxAgeMinutes  int64  `yaml:"max_age_minutes"`
	RotationSizeMB int64  `yaml:"rotation_size_mb"`
}

// Log 日志的配置.
type Log struct {
	Zap Zap `yaml:"zap"`
}

// VictoriaMetrics vm push的配置
type VictoriaMetrics struct {
	Mode                string `yaml:"mode"` // push 还是 pull
	PushAddr            string `yaml:"push_addr"`
	PushIntervalSeconds int64  `yaml:"push_interval_seconds"`
	PushBitsFlags       uint64 `yaml:"push_bits_flags"`
	PullServerAddr      string `yaml:"pull_server_addr"` // 监听的端口
	PullHTTPPath        string `yaml:"pull_http_path"`   // 默认 /metrics
}

// Metrics 监控的上报.
type Metrics struct {
	VictoriaMetrics VictoriaMetrics `yaml:"victoria_metrics"`
}

// Plugins 插件的配置.
type Plugins struct {
	Log     Log     `yaml:"log"`
	Metrics Metrics `yaml:"metrics"`
}

// Configs 整个进程的基本配置.
type Configs struct {
	Global  Global  `yaml:"global"`
	Server  Server  `yaml:"server"`
	Plugins Plugins `yaml:"plugins"`
}

//nolint:gochecknoglobals  // I just need it
var cfg Configs // todo: 未来考虑做成动态加载

// GetConfig 获取配置文件.
func GetConfig() *Configs {
	return &cfg
}

//nolint:gochecknoglobals // I just need it
var configFile = flag.String("config_file", "", "a yaml config file for server")

// LoadFromCmdLine 通过`-config_file=xxx`的命令行来加载配置.
func LoadFromCmdLine() error {
	if len(*configFile) == 0 {
		return fmt.Errorf("[%s]cmd line not set -config_file=xxx", debugs.SourceCodeLoc(1))
	}
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		return fmt.Errorf("[%s]-config_file=%s not exists", debugs.SourceCodeLoc(1), *configFile)
	}
	yamlData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return fmt.Errorf("[%s]-config_file=%s, read config error, err=%w",
			debugs.SourceCodeLoc(1), *configFile, err)
	}
	return LoadFromBytes(yamlData)
}

// LoadFromBytes 通过yaml文件内容加载配置.
func LoadFromBytes(yamlData []byte) error {
	if err := yaml.Unmarshal(yamlData, &cfg); err != nil {
		return fmt.Errorf("[%s]yaml decode error, err=%w",
			debugs.SourceCodeLoc(1), err)
	}
	modifyValues()
	return nil
}

func modifyValues() {
	if cfg.Global.ContainerName == "${container_name}" {
		cfg.Global.ContainerName = os.Getenv("POD_NAME")
	}
	if cfg.Global.LocalIP == "${local_ip}" {
		cfg.Global.LocalIP = stringsutil.GetContainerIP()
	}
}
