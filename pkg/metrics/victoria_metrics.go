package metrics

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	vm "github.com/ahfuzhang/metrics"
	"github.com/gin-gonic/gin"

	"github.com/ahfuzhang/cheaprpc/pkg/config"
	"github.com/ahfuzhang/cheaprpc/pkg/util/debugs"
)

const (
	defaultPushIntervalSeconds = 30
)

func initPushMode() error {
	cfg := config.GetConfig()
	metricCfg := cfg.Plugins.Metrics.VictoriaMetrics
	addr := metricCfg.PushAddr
	t, err := template.New("addr").Parse(addr)
	if err != nil {
		return fmt.Errorf("[%s]addr %s format error, err=%+v", debugs.SourceCodeLoc(1), addr, err)
	}
	addrData := &bytes.Buffer{}
	err = t.Execute(addrData, map[string]string{
		"App":    cfg.Server.App,
		"Server": cfg.Server.Server,
		"IP":     cfg.Global.LocalIP,
	})
	if err != nil {
		return fmt.Errorf("[%s]run template %s error, err=%+v", debugs.SourceCodeLoc(1), addr, err)
	}
	pushIntervalSeconds := metricCfg.PushIntervalSeconds
	if pushIntervalSeconds < 5 || pushIntervalSeconds > 60*5 {
		pushIntervalSeconds = defaultPushIntervalSeconds
	}
	url := addrData.String()
	var baseLabels string
	baseLabels, err = getBaseLabels()
	if err != nil {
		return debugs.WarpError(err, "getBaseLabels error")
	}
	err = vm.InitPushWithFlags(url, time.Duration(pushIntervalSeconds)*time.Second, baseLabels,
		metricCfg.PushBitsFlags) // vm.FlagOfPushProcessMetrics|vm.FlagOfForbideGzip)
	if err != nil {
		return debugs.WarpError(err, "vm.InitPush error")
	}
	return nil
}

// Init 初始化监控上报
func Init(g *gin.Engine) error {
	cfg := config.GetConfig()
	metricCfg := cfg.Plugins.Metrics.VictoriaMetrics
	if metricCfg.Mode == "push" || metricCfg.Mode == "both" {
		if err := initPushMode(); err != nil {
			return err
		}
	}
	if metricCfg.Mode == "pull" || metricCfg.Mode == "both" {
		if metricCfg.PullHTTPPath == "" || !strings.HasPrefix(metricCfg.PullHTTPPath, "/") {
			metricCfg.PullHTTPPath = "/metrics"
		}
		if g != nil {
			g.GET(metricCfg.PullHTTPPath, func(context *gin.Context) {
				vm.WritePrometheus(context.Writer, (metricCfg.PushBitsFlags&uint64(vm.FlagOfPushProcessMetrics)) != 0)
			})
		}
	}
	return nil
}

func getBaseLabels() (string, error) {
	t, err := template.New("base_label").Parse(
		`namespace="{{.Global.Namespace}}",` +
			`env_name="{{.Global.EnvName}}",` +
			`container_name="{{.Global.ContainerName}}",` +
			`local_ip="{{.Global.LocalIP}}",` +
			`app="{{.Server.App}}",` +
			`server="{{.Server.Server}}"`)
	if err != nil {
		return "", debugs.WarpError(err, "parse template error")
	}
	cfg := config.GetConfig()
	out := &bytes.Buffer{}
	err = t.Execute(out, cfg)
	if err != nil {
		log.Println("template execute error, err=", err.Error())
		return "", debugs.WarpError(err, "execute template error")
	}
	return out.String(), nil
}
