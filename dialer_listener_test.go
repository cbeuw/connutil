package connutil

import (
	"context"
	"io"
	"testing"
	"time"
)

func TestDial_Accept(t *testing.T) {
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

func TestPipeDialer_DialContext(t *testing.T) {
	d, l := DialerListener(1)
	a, _ := d.Dial("", "")
	ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
	_, err := d.DialContext(ctx, "", "")

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
}

func TestListener_Addr(t *testing.T) {
	_, l := DialerListener(1)
	if l.Addr() == nil {
		t.Error("listener's address shouldn't be nil")
	}
}
