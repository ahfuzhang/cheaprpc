package myeasyservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ahfuzhang/cheaprpc/pkg/netframework"

	"github.com/ahfuzhang/my_easy_service/pb"
)

var (
	etcdPath  = template.Must(template.New("path").Parse(`/cheaprpc/{{.Namespace}}/{{.EnvName}}/{{.App}}/{{.Server}}/{{.Addr}}`))
	etcdValue = template.Must(template.New("path").Parse(`namespace={{.Namespace}}
env_name={{.EnvName}}
app={{.App}}
server={{.Server}}
addr={{.Addr}}
services={{.Services}}
`))
)

type ServerConfig struct {
	Namespace string
	EnvName   string
	App       string
	Server    string
	Addr      string
	Services  string
}

func GetEtcdRegisterData() (string, string) {
	param := ServerConfig{
		Namespace: os.Getenv("namespace"),
		EnvName:   os.Getenv("env_name"),
		App:       os.Getenv("app"),
		Server:    os.Getenv("server"),
		Addr:      os.Getenv("addr"),
		Services:  os.Getenv("services"),
	}
	path := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := etcdPath.Execute(path, &param); err != nil {
		log.Fatalf("template error")
	}
	value := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := etcdValue.Execute(value, &param); err != nil {
		log.Fatalf("template error")
	}
	return path.String(), value.String()
}

func etcd1() error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"10.129.100.199:31433", "10.129.100.200:31433", "10.129.100.201:31433"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer client.Close()
	//
	lease := clientv3.NewLease(client)
	rsp, err := lease.Grant(context.Background(), 60)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
	path, value := GetEtcdRegisterData()
	resp, err := client.Put(ctx, path, value, clientv3.WithLease(rsp.ID))
	cancel()
	if err != nil {
		// handle error!
	}
	log.Println(resp)
	//
	c, _ := lease.KeepAlive(context.Background(), rsp.ID)
	for {
		ka := <-c
		log.Println("ttl:", ka.TTL)
		time.Sleep(time.Duration(1) * time.Second)
	}
	//
	return nil
}

func Register(framework netframework.Framework) {
	engine, ok := framework.GetServiceHandleRegister().(*gin.Engine)
	if !ok {
		log.Fatalf("framework.GetRegister().(*gin.Engine) fail")
	}

	engine.POST("/cheaprpc.ahfuzhang.my_easy_service.MyEasyService.GetEchoInfo", func(ctx *gin.Context) {
		body, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"code": 400,
				"msg":  fmt.Sprintf("read body error,err=%+v", err),
			})
			return
		}
		req := &pb.GetReq{}
		err = json.Unmarshal(body, req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"code": 400,
				"msg":  fmt.Sprintf("json decode error,err=%+v", err),
			})
			return
		}
		var rsp *pb.GetRsp
		rsp, err = Instance.GetEchoInfo(ctx, req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code": 500,
				"msg":  fmt.Sprintf("server error,err=%+v", err),
			})
			return
		}
		ctx.JSON(http.StatusOK, rsp)
	})
	engine.POST("http/request/get_echo_info", func(ctx *gin.Context) {
		body, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"code": 400,
				"msg":  fmt.Sprintf("read body error,err=%+v", err),
			})
			return
		}
		req := &pb.GetReq{}
		err = json.Unmarshal(body, req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"code": 400,
				"msg":  fmt.Sprintf("json decode error,err=%+v", err),
			})
			return
		}
		var rsp *pb.GetRsp
		rsp, err = Instance.GetEchoInfo(ctx, req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code": 500,
				"msg":  fmt.Sprintf("server error,err=%+v", err),
			})
			return
		}
		ctx.JSON(http.StatusOK, rsp)
	})
	engine.POST("/cheaprpc.ahfuzhang.my_easy_service.MyEasyService.Save", func(ctx *gin.Context) {
		body, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"code": 400,
				"msg":  fmt.Sprintf("read body error,err=%+v", err),
			})
			return
		}
		req := &pb.SaveReq{}
		err = json.Unmarshal(body, req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"code": 400,
				"msg":  fmt.Sprintf("json decode error,err=%+v", err),
			})
			return
		}
		var rsp *pb.SaveRsp
		rsp, err = Instance.Save(ctx, req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code": 500,
				"msg":  fmt.Sprintf("server error,err=%+v", err),
			})
			return
		}
		ctx.JSON(http.StatusOK, rsp)
	})

}
