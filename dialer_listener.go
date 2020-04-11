package connutil

import (
	"context"
	"net"
	"sync/atomic"
)

// Dialer is an interface such that the net.Dialer type implements this. This interface type can be used
// as a drop-in replacement for type bindings where methods of net.Dialer are called.
type Dialer interface {
	Dial(network, address string) (net.Conn, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// PipeDialer implements the Dialer interface, which means it has the same method signatures as net.Dialer
type PipeDialer struct {
	// PipeBufferSize specifies the limit on the underlying buffer size. Default (0) means unlimited.
	BufferSizeLimit int
	peer            *PipeListener
}

// Dial returns one end of the pipe, with the other end to be obtained from Accept.
// network and address arguments don't do anything
func (d *PipeDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext returns one end of the pipe, with the other end to be obtained from Accept.
//
// network and address don't do anything. It can timeout or be cancelled using ctx.
func (d *PipeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if atomic.LoadUint32(&d.peer.closed) == 1 {
		return nil, ErrListenerClosed
	}
	a, b := LimitedAsyncPipe(d.BufferSizeLimit)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d.peer.incomingConn <- b:
		return a, nil
	}
}

// PipeListener is a net.Listener that accepts connections dialed from the corresponding PipeDialer
type PipeListener struct {
	incomingConn chan net.Conn
	closed       uint32
}

// Accept implements Listener.Accept(). It returns one end of the pipe, with the other end obtained from Dial
func (l *PipeListener) Accept() (net.Conn, error) {
	if atomic.LoadUint32(&l.closed) == 1 {
		return nil, ErrListenerClosed
	} else {
		return <-l.incomingConn, nil
	}
}

// Close implements Listener.Close()
func (l *PipeListener) Close() error {
	atomic.StoreUint32(&l.closed, 1)
	return nil
}

// Addr returns a fake address
func (l *PipeListener) Addr() net.Addr {
	return fakeAddr{}
}

// DialerListener returns a pair of PipeDialer and PipeListener. By calling
// PipeDialer.Dial and then PipeListener.Accept, you can get two ends of an asynchronous pipe,
// as if obtained through AsyncPipe()
//
// backlog specifies the amount of Dial calls you can make without making corresponding Accept calls on listener before
// Dial calls start blocking
func DialerListener(backlog int) (Dialer, net.Listener) {
	l := &PipeListener{
		incomingConn: make(chan net.Conn, backlog),
		closed:       0,
	}
	d := &PipeDialer{peer: l}
	return d, l
}
