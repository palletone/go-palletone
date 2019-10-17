package comm

import (
	"net"
	"strings"
)

func GetInternalIp() string {
	conn,err := net.Dial("udp","8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	internalIp := conn.LocalAddr().String()
	idx := strings.LastIndex(internalIp,":")
	return internalIp[0:idx]
}
