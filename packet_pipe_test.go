package connutil

import (
	"bytes"
	"io"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestPacketPipeConn_WriteTo(t *testing.T) {
	testData := make([]byte, 128)
	t.Run("large write", func(t *testing.T) {
		_, w := LimitedAsyncPacketPipe(1)
		_, err := w.WriteTo(testData, nil)
		if err != ErrWriteToLarge {
			t.Errorf("expecting %v, got %v", ErrWriteToLarge, err)
		}
	})
	t.Run("write after close", func(t *testing.T) {
		_, w := AsyncPacketPipe()
		err := w.Close()
		if err != nil {
			t.Error(err)
		}
		_, err = w.WriteTo(testData, nil)
		if err != io.ErrClosedPipe {
			t.Errorf("expecting %v, got %v", io.ErrClosedPipe, err)
		}
	})
	t.Run("write after deadline passed", func(t *testing.T) {
		_, w := AsyncPacketPipe()
		err := w.SetWriteDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, err = w.WriteTo(testData, nil)
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
	t.Run("recover from write deadline", func(t *testing.T) {
		_, w := AsyncPacketPipe()
		err := w.SetWriteDeadline(time.Now().Add(-1 * time.Second))
		_, err = w.WriteTo(testData, nil)
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
		err = w.SetWriteDeadline(time.Time{})
		if err != nil {
			t.Error(err)
		}
		_, err = w.WriteTo(testData, nil)
		if err != nil {
			t.Error("cannot write after removing deadline")
		}
	})
}

func TestPacketPipeConn_ReadFrom(t *testing.T) {
	makeBuf := func() []byte {
		return make([]byte, 128)
	}
	t.Run("read after close", func(t *testing.T) {
		r, _ := AsyncPacketPipe()
		err := r.Close()
		if err != nil {
			t.Error(err)
		}
		_, _, err = r.ReadFrom(makeBuf())
		if err != io.ErrClosedPipe {
			t.Errorf("expecting %v, got %v", io.ErrClosedPipe, err)
		}
	})
	t.Run("read block then timeout", func(t *testing.T) {
		r, _ := AsyncPacketPipe()
		done := make(chan struct{})
		go func() {
			_, _, _ = r.ReadFrom(make([]byte, 1))
			done <- struct{}{}
		}()

		_ = r.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		select {
		case <-done:
			return
		case <-time.After(1 * time.Second):
			t.Error("Read did not unblock after deadline has passed")
		}
	})

	t.Run("read after deadline passed", func(t *testing.T) {
		r, _ := AsyncPacketPipe()
		err := r.SetReadDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, _, err = r.ReadFrom(makeBuf())
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
	t.Run("recover from read deadline", func(t *testing.T) {
		r, w := AsyncPacketPipe()
		err := r.SetReadDeadline(time.Now().Add(-1 * time.Second))
		_, _, err = r.ReadFrom(makeBuf())
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
		err = r.SetReadDeadline(time.Time{})
		if err != nil {
			t.Error(err)
		}
		_, _ = w.WriteTo(makeBuf(), nil)
		_, _, err = r.ReadFrom(makeBuf())
		if err != nil {
			t.Error("cannot read after removing deadline")
		}
	})
}

func TestPacketPipeConn_ReadWrite(t *testing.T) {
	testData := make([]byte, 128)
	t.Run("simple read write", func(t *testing.T) {
		r, w := AsyncPacketPipe()
		_, err := w.WriteTo(testData, nil)
		if err != nil {
			t.Error(err)
		}

		receiveBuf := make([]byte, len(testData))
		n, _, err := r.ReadFrom(receiveBuf)
		if n != len(testData) {
			t.Error("short read")
		}
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("read write with deadline", func(t *testing.T) {
		r, w := AsyncPacketPipe()
		_ = w.SetWriteDeadline(time.Now().Add(1 * time.Second))
		_ = r.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, err := w.WriteTo(testData, nil)
		if err != nil {
			t.Error("write timed out")
		}

		_, _, err = r.ReadFrom(make([]byte, len(testData)))
		if err != nil {
			t.Error("read timed out")
		}
	})
	t.Run("read write after deadline passed", func(t *testing.T) {
		c, _ := AsyncPacketPipe()
		err := c.SetDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, err = c.WriteTo(testData, nil)
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}

		_, _, err = c.ReadFrom(make([]byte, len(testData)))
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
	t.Run("packet integrity", func(t *testing.T) {
		r, w := AsyncPacketPipe()
		_, _ = w.WriteTo(testData, nil)

		receiveBuf := make([]byte, len(testData)*2)
		_, _, err := r.ReadFrom(receiveBuf[:len(testData)-1])
		if err != io.ErrShortBuffer {
			t.Errorf("expecting %v, got %v", io.ErrShortBuffer, err)
		}
		n, _, err := r.ReadFrom(receiveBuf)
		if n != len(testData) {
			t.Errorf("wrong amount of bytes read: %v", n)
		}
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(testData, receiveBuf[:len(testData)]) {
			t.Error("read wrong data")
		}
	})
}

func TestAsyncPacketPipe_Addrs(t *testing.T) {
	p, _ := AsyncPacketPipe()
	if p.LocalAddr() == nil {
		t.Error("LocalAddr shouldn't return null pointer")
	}
	p.LocalAddr().Network()
	p.LocalAddr().String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestLimitedAsyncPacketPipe(t *testing.T) {
	testData := make([]byte, 1<<16)
	rand.Read(testData)
	t.Run("large write", func(t *testing.T) {
		var packetsOnWire uint32
		const bufferSize = 128
		w, r := LimitedAsyncPacketPipe(bufferSize)
		go func() {
			contentLen := len(testData)
			var n int
			for n < contentLen {
				atomic.AddUint32(&packetsOnWire, 1)
				toWrite := rand.Intn(bufferSize) + 1
				written, _ := w.WriteTo(testData[n:min(n+toWrite, contentLen)], nil)
				n += written
			}
		}()

		readBuf := make([]byte, len(testData))
		var n int
		for n < len(testData) {
			atomic.AddUint32(&packetsOnWire, ^uint32(0))
			i, _, err := r.ReadFrom(readBuf[n:])
			if err != nil {
				t.Error(err)
			}
			n += i
		}
		if packetsOnWire != 0 {
			t.Error("number of packets sent doesn't equal to number received")
		}
		if !bytes.Equal(readBuf, testData) {
			t.Error("wrong data read")
		}
	})
}
