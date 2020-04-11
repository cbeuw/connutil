package connutil

import (
	"net"
	"time"
)

type PacketPipe struct {
	writeEnd *bufferedPacketPipe
	readEnd  *bufferedPacketPipe
}

func (conn *PacketPipe) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = conn.Read(p)
	addr = conn.RemoteAddr()
	return
}

func (conn *PacketPipe) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n, err = conn.Write(p)
	return
}

func (conn *PacketPipe) Read(p []byte) (n int, err error) {
	n, err = conn.readEnd.Read(p)
	return n, err
}

func (conn *PacketPipe) Write(p []byte) (n int, err error) {
	n, err = conn.writeEnd.Write(p)
	return
}

func (conn *PacketPipe) Close() error {
	conn.writeEnd.Close()
	conn.readEnd.Close()
	return nil
}

func (conn *PacketPipe) SetReadDeadline(t time.Time) error {
	conn.readEnd.SetReadDeadline(t)
	return nil
}
func (conn *PacketPipe) SetWriteDeadline(t time.Time) error {
	conn.writeEnd.SetWriteDeadline(t)
	return nil
}

func (conn *PacketPipe) SetDeadline(t time.Time) error {
	_ = conn.SetReadDeadline(t)
	_ = conn.SetWriteDeadline(t)
	return nil
}

func (conn *PacketPipe) LocalAddr() net.Addr  { return fakeAddr{} }
func (conn *PacketPipe) RemoteAddr() net.Addr { return fakeAddr{} }

// AsyncPipe is an in-memory, full-duplex, packet-oriented pipe with both ends implementing net.Conn and net.PacketConn
// interfaces.
//
// It is buffered, asynchronous and safe for concurrent use.
// Read calls will block until data becomes available by writing to the other end,
// Write calls on either end will never block, but it will panic if the buffer becomes too large for the memory.
func AsyncPacketPipe() (*PacketPipe, *PacketPipe) {
	return LimitedAsyncPacketPipe(0)
}

// LimitedAsyncPipe is similar to AsyncPipe, but Write calls will block if the buffer size grows larger than
// bufferSizeLimit. Consuming data through Read calls on the other end will unblock the Write call. If the size
// of the packet provided to a Write call is larger than bufferSizeLimit, it will return ErrWriteToLarge.
func LimitedAsyncPacketPipe(bufferSizeLimit int) (*PacketPipe, *PacketPipe) {
	LtoR := newBufferedPacketPipe(bufferSizeLimit)
	RtoL := newBufferedPacketPipe(bufferSizeLimit)
	a := &PacketPipe{
		writeEnd: LtoR,
		readEnd:  RtoL,
	}
	b := &PacketPipe{
		writeEnd: RtoL,
		readEnd:  LtoR,
	}
	return a, b
}
