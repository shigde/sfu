package rtp

import "net"

type UdpConnection struct {
	conn        *net.UDPConn
	port        int
	payloadType uint8
}

type UdpShare struct {
	Audio UdpShareInfo
	Video UdpShareInfo
}

type UdpShareInfo struct {
	Port        int
	PayloadType uint8
}
