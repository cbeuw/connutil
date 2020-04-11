package connutil

import (
	"io"
	"testing"
	"time"
)

func TestNullReader_Read(t *testing.T) {
	n := new(NullReader)
	recvBuf := make([]byte, 128)
	_, err := io.ReadFull(n, recvBuf)
	if err != nil {
		t.Error(err)
	}
}

func TestNullReader_WriteTo(t *testing.T) {
	n := new(NullReader)

	const readFor = 500 * time.Millisecond
	dump := Discard()
	_ = dump.SetDeadline(time.Now().Add(readFor))

	done := make(chan struct{})
	go func() {
		_, _ = n.WriteTo(dump)
		done <- struct{}{}
	}()
	select {
	case <-done:
		return
	case <-time.After(readFor + 100*time.Millisecond):
		t.Error("not finished after deadline")
	}
}
