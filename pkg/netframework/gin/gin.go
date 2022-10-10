// Package gin 使用gin框架作为cheaprpc的网络框架
package gin

import (
	thegin "github.com/gin-gonic/gin"

	"github.com/ahfuzhang/cheaprpc/pkg/netframework"
)

type GinFramework struct {
	engine *thegin.Engine
}

func NewGinFramework() (netframework.Framework, error) {
	out := &GinFramework{
		engine: thegin.New(),
	}
	return out, nil
}

func (f *GinFramework) Start(addr string) error {
	go func() {
		f.engine.Run(addr)
	}()
	return nil
}

func (f *GinFramework) GetRegister() interface{} {
	return f.engine
}
