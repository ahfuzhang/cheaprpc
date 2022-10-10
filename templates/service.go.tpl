package {{.ServicePackage}}

import (
	"context"
	"fmt"

	"{{.Import}}"
)

type {{.Service}}Imp struct {
}

var Instance = &{{.Service}}Imp{
    // todo: add init field if you need
}

func init(){
   // todo: init Instance if you need
}

{{$Service := .Service}}
{{- range $item := .Methods }}
func (s *{{$Service}}Imp) {{$item.Method}}(ctx context.Context, req *pb.{{$item.Req}}) (*pb.{{$item.Rsp}}, error) {
    // TODO: add biz code here
	return nil, fmt.Errorf("not implement")
}

{{- end }}

