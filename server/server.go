package server

import (
	"net"
)

const (
	port   = "53535"
	hostIP = "127.0.0.1"
)

func Serve() error {
	udpAddr, err := net.ResolveUDPAddr("udp", hostIP+":"+port)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	for {
		var buf [512]byte
		_, addr, err := conn.ReadFromUDP(buf[:])
		if err != nil {
			return err
		}
		go handlePacket(conn, addr, buf)
	}

	//return nil
}
