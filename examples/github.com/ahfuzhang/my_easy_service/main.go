package main

import (
	"fmt"
	golog "log"
	"os"
	"os/signal"
	"syscall"

	thegin "github.com/gin-gonic/gin"

	"github.com/ahfuzhang/cheaprpc/pkg/config"
	log "github.com/ahfuzhang/cheaprpc/pkg/log/zap"
	"github.com/ahfuzhang/cheaprpc/pkg/metrics"
	"github.com/ahfuzhang/cheaprpc/pkg/netframework"
	"github.com/ahfuzhang/cheaprpc/pkg/netframework/gin"
	"github.com/ahfuzhang/my_easy_service/internal/services/myeasyservice"
	"github.com/ahfuzhang/my_easy_service/internal/services/myeasyservice2"

	_ "go.uber.org/automaxprocs"
)

func NewFramework(name string) (netframework.Framework, error) {
	switch name {
	case "gin":
		return gin.NewGinFramework()
	default:
		log.Fatalf("not support net framework:" + name)
		return nil, fmt.Errorf("not support net framework:%+v", name)
	}
}

func main() {
	golog.SetFlags(golog.Lshortfile | golog.LstdFlags)
	if err := config.LoadFromCmdLine(); err != nil {
		golog.Fatalln(err)
	}
	cfg := config.GetConfig()
	zapCfg := &cfg.Plugins.Log.Zap
	if err := log.Init(cfg.Server.App, cfg.Server.Server, zapCfg); err != nil {
		golog.Fatalln(err)
	}
	// logrus.SetLevel(logrus.TraceLevel)
	// logrus.SetReportCaller(true)
	// log.SetFlags(log.Lshortfile | log.LstdFlags)
	// addr := ":8080"
	// if len(os.Args) >= 2 {
	// 	addr = os.Args[1]
	// }

	// start
	addrs := make(map[string]netframework.Framework)
	services := make(map[string]netframework.Framework)
	var err error
	for _, item := range cfg.Server.Service {
		framework, has := addrs[item.Addr]
		if has {
			services[item.Name] = framework
			continue
		}
		frameworkType := "gin"
		// if len(os.Args) >= 3 {
		// 	frameworkType = os.Args[2]
		// }
		framework, err = NewFramework(frameworkType)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		addrs[item.Addr] = framework
		if err = framework.Start(item.Addr); err != nil {
			log.Fatalf("framework.Start fail, %+v", err)
		}
		services[item.Name] = framework
	}
	// todo: add register code here
	myeasyservice.Register(services["name1"])
	myeasyservice2.Register(services["name2"])
	// 注册监控上报  // todo: 很丑陋
	if err = metrics.Init(services[cfg.Server.Service[0].Name].GetServiceHandleRegister().(*thegin.Engine)); err != nil {
		golog.Fatalln(err)
	}
	// todo: 注册到ETCD

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	// todo: 有个总的context
}
