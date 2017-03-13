package beacon

import "net"

type sender struct {
	socket *net.UDPConn
}

func newSender(address string, port int) (s *sender, err error) {
	sock, err := net.DialUDP(
		"udp4",
		&net.UDPAddr{
			IP:   net.ParseIP(address),
			Port: 0},
		&net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: port,
		},
	)
	if err != nil {
		return
	}

	s = &sender{
		socket: sock,
	}

	return

}

func (s *sender) send(payload []byte) {
	s.socket.Write(payload)
}

func (s *sender) close() {
	s.socket.Close()
}
