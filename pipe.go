package connutil

import (
	"net"
	"time"
)

type pipeConn struct {
	writeEnd *bufferedPipe
	readEnd  *bufferedPipe
}

func (conn *pipeConn) Read(b []byte) (int, error)  { return conn.readEnd.Read(b) }
func (conn *pipeConn) Write(b []byte) (int, error) { return conn.writeEnd.Write(b) }
func (conn *pipeConn) Close() error {
	conn.writeEnd.Close()
	conn.readEnd.Close()
	return nil
}

func (conn *pipeConn) SetReadDeadline(t time.Time) error { conn.readEnd.SetReadDeadline(t); return nil }
func (conn *pipeConn) SetWriteDeadline(t time.Time) error {
	conn.writeEnd.SetWriteDeadline(t)
	return nil
}

func (conn *pipeConn) SetDeadline(t time.Time) error {
	_ = conn.SetReadDeadline(t)
	_ = conn.SetWriteDeadline(t)
	return nil
}

func (conn *pipeConn) LocalAddr() net.Addr  { return fakeAddr{} }
func (conn *pipeConn) RemoteAddr() net.Addr { return fakeAddr{} }

// AsyncPipe is an in-memory, full-duplex pipe with both ends implementing net.Conn interface.
//
// It is a drop-in replacement of net.Pipe, but buffered, asynchronous and safe for concurrent use.
// Read calls will block until data becomes available by writing to the other end,
// Write calls on either end will never block, but it will panic if the buffer becomes too large for the memory.
func AsyncPipe() (net.Conn, net.Conn) {
	return LimitedAsyncPipe(0)
}

// LimitedAsyncPipe is similar to AsyncPipe, but Write calls will block if the buffer size grows larger than
// bufferSizeLimit. Consuming data through Read calls on the other end will unblock the Write call.
func LimitedAsyncPipe(bufferSizeLimit int) (net.Conn, net.Conn) {
	LtoR := newBufferedPipe(bufferSizeLimit)
	RtoL := newBufferedPipe(bufferSizeLimit)
	a := &pipeConn{
		writeEnd: LtoR,
		readEnd:  RtoL,
	}
	b := &pipeConn{
		writeEnd: RtoL,
		readEnd:  LtoR,
	}
	return a, b
}
