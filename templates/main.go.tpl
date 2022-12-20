package main

import (
    "fmt"
	"log"
	"os"
	"os/signal"
    "syscall"

	"github.com/ahfuzhang/cheaprpc/pkg/netframework"
	"github.com/ahfuzhang/cheaprpc/pkg/netframework/gin"

	_ "go.uber.org/automaxprocs"

{{$Package := .Package}}
{{- range $item := .Services }}
	"{{$Package}}/internal/services/{{.}}"
{{- end }}
)

func NewFramework(name string) (netframework.Framework, error) {
	switch name {
	case "gin":
		return gin.NewGinFramework()
	default:
		log.Fatalln("not support net framework:" + name)
		return nil, fmt.Errorf("not support net framework:%+v", name)
	}
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	addr := ":8080"
	if len(os.Args) >= 2 {
		addr = os.Args[1]
	}
	frameworkType := "gin"
	if len(os.Args) >= 3 {
		frameworkType = os.Args[2]
	}
	framework, err := NewFramework(frameworkType)
	if err != nil {
		log.Fatalln(err)
	}
	// todo: add register code here
{{- range $item := .Services }}
	{{.}}.Register(framework)
{{- end }}

	// start
	if err = framework.Start(addr); err != nil {
		log.Fatalln("start fail:", err.Error())
	}
	c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
    <-c
}
