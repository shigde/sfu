package rtp

import "net"

type UdpConnection struct {
	conn        *net.UDPConn
	port        int
	payloadType uint8
}
