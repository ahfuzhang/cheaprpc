package netframework

type Framework interface {
	Start(addr string) error
	GetRegister() interface{}
}
