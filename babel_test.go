package connutil

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
	"time"
)

func TestBabelConn_Write(t *testing.T) {
	testData := make([]byte, 128)
	randReader := rand.New(rand.NewSource(0))
	t.Run("simple write", func(t *testing.T) {
		_, err := Babel(randReader).Write(testData)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("write after close", func(t *testing.T) {
		babel := Babel(rand.New(randReader))
		err := babel.Close()
		if err != nil {
			t.Error(err)
		}
		_, err = babel.Write(testData)
		if err != io.ErrClosedPipe {
			t.Errorf("expecting %v, got %v", io.ErrClosedPipe, err)
		}
	})
	t.Run("write after deadline passed", func(t *testing.T) {
		babel := Babel(rand.New(randReader))
		err := babel.SetWriteDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, err = babel.Write(testData)
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
	t.Run("recover from deadline", func(t *testing.T) {
		babel := Babel(rand.New(randReader))
		_ = babel.SetWriteDeadline(time.Now().Add(-1 * time.Second))
		_, _ = babel.Write(testData)
		err := babel.SetWriteDeadline(time.Time{})
		if err != nil {
			t.Error(err)
		}
		_, err = babel.Write(testData)
		if err != nil {
			t.Error("cannot write after removing deadline")
		}
	})
}

func TestBabelConn_Read(t *testing.T) {
	makeBuf := func() []byte {
		return make([]byte, 128)
	}
	randReader := rand.New(rand.NewSource(0))

	t.Run("simple read", func(t *testing.T) {
		exp := rand.New(rand.NewSource(0))
		b := Babel(rand.New(rand.NewSource(0)))
		babelBuf := makeBuf()
		_, err := io.ReadFull(b, babelBuf)
		if err != nil {
			t.Error(err)
		}
		expBuf := makeBuf()
		_, _ = io.ReadFull(exp, expBuf)
		if !bytes.Equal(babelBuf, expBuf) {
			t.Error("babel didn't read correct data")
		}
	})
	t.Run("read after close", func(t *testing.T) {
		b := Babel(randReader)
		err := b.Close()
		if err != nil {
			t.Error(err)
		}
		_, err = b.Read(makeBuf())
		if err != io.ErrClosedPipe {
			t.Errorf("expecting %v, got %v", io.ErrClosedPipe, err)
		}
	})
	t.Run("read block then timeout", func(t *testing.T) {
		b := Babel(randReader)
		done := make(chan struct{})
		go func() {
			_, _ = b.Read(make([]byte, 1))
			done <- struct{}{}
		}()

		_ = b.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		select {
		case <-done:
			return
		case <-time.After(1 * time.Second):
			t.Error("Read did not unblock after deadline has passed")
		}
	})

	t.Run("read after deadline passed", func(t *testing.T) {
		b := Babel(randReader)
		err := b.SetReadDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, err = b.Read(makeBuf())
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
	t.Run("recover from read deadline", func(t *testing.T) {
		b := Babel(randReader)
		err := b.SetReadDeadline(time.Now().Add(-1 * time.Second))
		_, err = b.Read(makeBuf())
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
		err = b.SetReadDeadline(time.Time{})
		if err != nil {
			t.Error(err)
		}
		_, err = b.Read(makeBuf())
		if err != nil {
			t.Error("cannot read after removing deadline")
		}
	})
}

func TestBabelConn_SetDeadline(t *testing.T) {
	testData := make([]byte, 128)
	b := Babel(rand.New(rand.NewSource(0)))
	_ = b.SetDeadline(time.Now().Add(1 * time.Second))
	_, err := b.Write(testData)
	if err != nil {
		t.Error("write timed out")
	}

	_, err = b.Read(make([]byte, len(testData)))
	if err != nil {
		t.Error("read timed out")
	}
}

func TestBabelConn_Addrs(t *testing.T) {
	b := Babel(rand.New(rand.NewSource(0)))
	if b.LocalAddr() == nil {
		t.Error("LocalAddr shouldn't return null pointer")
	}
	if b.RemoteAddr() == nil {
		t.Error("RemoteAddr shouldn't return null pointer")
	}
}
