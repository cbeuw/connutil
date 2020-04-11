package connutil

import (
	"context"
	"io"
	"testing"
	"time"
)

func TestDialerListener_Stream(t *testing.T) {
	d, l := DialerListener(1)
	a, err := d.Dial("", "")
	if err != nil {
		t.Error(err)
	}
	b, err := l.Accept()
	if err != nil {
		t.Error(err)
	}

	_, err = a.Write(make([]byte, 16))
	if err != nil {
		t.Error(err)
	}
	_, err = io.ReadFull(b, make([]byte, 16))
	if err != nil {
		t.Error(err)
	}
}

func TestDialerListener_Packet(t *testing.T) {
	d, l := DialerListener(1)
	a, err := d.Dial("udp", "")
	if err != nil {
		t.Error(err)
	}
	b, err := l.ListenPacket("", "")
	if err != nil {
		t.Error(err)
	}

	_, err = b.WriteTo(make([]byte, 16), nil)
	if err != nil {
		t.Error(err)
	}

	_, err = a.Read(make([]byte, 16))
	if err != nil {
		t.Error(err)
	}
}

func TestPipeDialer_DialContext(t *testing.T) {
	d, l := DialerListener(1)

	a, _ := d.Dial("", "")
	ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
	_, err := d.DialContext(ctx, "tcp", "")
	if err != ctx.Err() {
		t.Errorf("expcting timeout, got %v", ctx.Err())
	}

	_, err = d.DialContext(ctx, "udp", "")
	if err != ctx.Err() {
		t.Errorf("expcting timeout, got %v", ctx.Err())
	}

	b, err := l.Accept()
	if err != nil {
		t.Errorf("should have accepted the dialed conn: %v", err)
	}

	_, err = a.Write(make([]byte, 16))
	if err != nil {
		t.Error(err)
	}
	_, err = io.ReadFull(b, make([]byte, 16))
	if err != nil {
		t.Error(err)
	}
}

func TestListener_Close(t *testing.T) {
	d, l := DialerListener(1)
	err := l.Close()
	if err != nil {
		t.Error(err)
	}
	_, err = d.Dial("", "")
	if err != ErrListenerClosed {
		t.Errorf("expecting %v, got %v", ErrListenerClosed, err)
	}
	_, err = l.Accept()
	if err != ErrListenerClosed {
		t.Errorf("expecting %v, got %v", ErrListenerClosed, err)
	}

	_, err = l.ListenPacket("", "")
	if err != ErrListenerClosed {
		t.Errorf("expecting %v, got %v", ErrListenerClosed, err)
	}
}

func TestListener_Addr(t *testing.T) {
	_, l := DialerListener(1)
	if l.Addr() == nil {
		t.Error("listener's address shouldn't be nil")
	}
}
