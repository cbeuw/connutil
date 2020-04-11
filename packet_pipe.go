package connutil

import (
	"net"
	"time"
)

type packetPipeConn struct {
	writeEnd *bufferedPacketPipe
	readEnd  *bufferedPacketPipe
}

func (conn *packetPipeConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = conn.readEnd.ReadFrom(p)
	return n, fakeAddr{}, err
}

func (conn *packetPipeConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n, err = conn.writeEnd.WriteTo(p)
	return
}

func (conn *packetPipeConn) Close() error {
	conn.writeEnd.Close()
	conn.readEnd.Close()
	return nil
}

func (conn *packetPipeConn) SetReadDeadline(t time.Time) error {
	conn.readEnd.SetReadDeadline(t)
	return nil
}
func (conn *packetPipeConn) SetWriteDeadline(t time.Time) error {
	conn.writeEnd.SetWriteDeadline(t)
	return nil
}

func (conn *packetPipeConn) SetDeadline(t time.Time) error {
	_ = conn.SetReadDeadline(t)
	_ = conn.SetWriteDeadline(t)
	return nil
}

func (conn *packetPipeConn) LocalAddr() net.Addr { return fakeAddr{} }

// AsyncPipe is an in-memory, full-duplex, packet-oriented pipe with both ends implementing net.PacketConn interface.
//
// It is buffered, asynchronous and safe for concurrent use.
// ReadFrom calls will block until data becomes available by writing to the other end,
// WriteTo calls on either end will never block, but it will panic if the buffer becomes too large for the memory.
func AsyncPacketPipe() (net.PacketConn, net.PacketConn) {
	return LimitedAsyncPacketPipe(0)
}

// LimitedAsyncPipe is similar to AsyncPipe, but WriteTo calls will block if the buffer size grows larger than
// bufferSizeLimit. Consuming data through ReadFrom calls on the other end will unblock the WriteTo call. If the size
// of the packet provided to a WriteTo call is larger than bufferSizeLimit, it will return ErrWriteToLarge.
func LimitedAsyncPacketPipe(bufferSizeLimit int) (net.PacketConn, net.PacketConn) {
	LtoR := newBufferedPacketPipe(bufferSizeLimit)
	RtoL := newBufferedPacketPipe(bufferSizeLimit)
	a := &packetPipeConn{
		writeEnd: LtoR,
		readEnd:  RtoL,
	}
	b := &packetPipeConn{
		writeEnd: RtoL,
		readEnd:  LtoR,
	}
	return a, b
}
