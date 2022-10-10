package myeasyservice

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"

    "github.com/ahfuzhang/cheaprpc/pkg/netframework"
    "github.com/ahfuzhang/my_easy_service/pb"
)

func Register(framework netframework.Framework){
    engine, ok := framework.GetRegister().(*gin.Engine)
    if !ok{
        log.Fatalf("framework.GetRegister().(*gin.Engine) fail")
    }

    engine.POST("/cheaprpc.ahfuzhang.my_easy_service.MyEasyService.GetEchoInfo", func(ctx *gin.Context){
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
    engine.POST("http/request/get_echo_info", func(ctx *gin.Context){
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
    engine.POST("/cheaprpc.ahfuzhang.my_easy_service.MyEasyService.Save", func(ctx *gin.Context){
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
