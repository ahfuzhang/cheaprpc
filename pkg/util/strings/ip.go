package strings

import (
	"net"
	"os"
)

// InterfaceAddrs 获取本机IP地址
func InterfaceAddrs() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// GetContainerIP 获取容器的IP
func GetContainerIP() string {
	ip := os.Getenv("POD_IP") // for TKEX env only
	if len(ip) == 0 {
		return InterfaceAddrs()
	}
	return ip
}
