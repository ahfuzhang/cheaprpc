package netframework

type Framework interface {
	Start(addr string) error
	GetServiceHandleRegister() interface{}
}
