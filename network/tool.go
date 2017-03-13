package network

import (
	"fmt"
	"net"
)

func GetFreePortOnInterface(iface string) (port int, err error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", iface))
	if err != nil {
		return
	}
	defer l.Close()
	port = l.Addr().(*net.TCPAddr).Port
	return
}
