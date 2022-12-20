// Package cheaprpc 具体的服务引用此框架的代码入口.
package cheaprpc

// Server 基于cheaprpc框架的服务类.
type Server struct {
}

// NewServer 封装服务处理的各个细节.
func NewServer() *Server {
	return &Server{}
}
