module {{.Package}}

go 1.17

replace (
	github.com/ahfuzhang/cheaprpc v0.0.1 => ../../../../
)

require (
	github.com/gin-gonic/gin v1.8.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.0
	github.com/jhump/protoreflect v1.13.0
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.28.0
	github.com/ahfuzhang/cheaprpc v0.0.1
)
