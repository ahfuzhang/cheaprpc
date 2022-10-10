package {{.ServicePackage}}

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"

    "github.com/ahfuzhang/cheaprpc/pkg/netframework"
    "{{.Import}}"
)

func Register(framework netframework.Framework){
    engine, ok := framework.GetRegister().(*gin.Engine)
    if !ok{
        log.Fatalf("framework.GetRegister().(*gin.Engine) fail")
    }
{{ with .Methods }}
{{- range . }}
    engine.POST("{{.Path}}", func(ctx *gin.Context){
        body, err := ctx.GetRawData()
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]interface{}{
                "code": 400,
                "msg":  fmt.Sprintf("read body error,err=%+v", err),
            })
            return
        }
        req := &pb.{{.Req}}{}
        err = json.Unmarshal(body, req)
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]interface{}{
                "code": 400,
                "msg":  fmt.Sprintf("json decode error,err=%+v", err),
            })
            return
        }
        var rsp *pb.{{.Rsp}}
        rsp, err = Instance.{{.Method}}(ctx, req)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
                "code": 500,
                "msg":  fmt.Sprintf("server error,err=%+v", err),
            })
            return
        }
        ctx.JSON(http.StatusOK, rsp)
    })
{{- end }}
{{ end }}
}
