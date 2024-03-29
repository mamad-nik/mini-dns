package server

import (
	"fmt"
	"net"
)

func handlePacket(conn *net.UDPConn, addr *net.UDPAddr, buf [512]byte) {
	fmt.Println(addr, string(buf[:]))
	_, err := conn.WriteToUDP(buf[:], addr)
	if err != nil {
		conn.WriteToUDP([]byte(err.Error()), addr)
		return
	}
}
