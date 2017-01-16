package emitter

import (
	"net"
)

type UDPEmitter struct {
	udpAddr *net.UDPAddr
	udpConn net.PacketConn
}

func NewUdpEmitter(remoteAddr string) (*UDPEmitter, error) {
	addr, err := net.ResolveUDPAddr("udp4", remoteAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenPacket("udp4", "")
	if err != nil {
		return nil, err
	}

	emitter := &UDPEmitter{udpAddr: addr, udpConn: conn}
	return emitter, nil
}

func (e *UDPEmitter) Emit(data []byte) error {
	_, err := e.udpConn.WriteTo(data, e.udpAddr)
	return err
}

func (e *UDPEmitter) Close() {
	e.udpConn.Close()
}

func (e *UDPEmitter) Address() net.Addr {
	return e.udpConn.LocalAddr()
}
