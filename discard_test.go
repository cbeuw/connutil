package connutil

import (
	"io"
	"testing"
	"time"
)

func TestDiscardConn_Write(t *testing.T) {
	testData := make([]byte, 128)
	t.Run("simple write", func(t *testing.T) {
		_, err := Discard().Write(testData)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("write after close", func(t *testing.T) {
		discard := Discard()
		err := discard.Close()
		if err != nil {
			t.Error(err)
		}
		_, err = discard.Write(testData)
		if err != io.ErrClosedPipe {
			t.Errorf("expecting %v, got %v", io.ErrClosedPipe, err)
		}
	})
	t.Run("write after deadline passed", func(t *testing.T) {
		discard := Discard()
		err := discard.SetWriteDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, err = discard.Write(testData)
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
	t.Run("recover from deadline", func(t *testing.T) {
		discard := Discard()
		_ = discard.SetWriteDeadline(time.Now().Add(-1 * time.Second))
		_, _ = discard.Write(testData)
		err := discard.SetWriteDeadline(time.Time{})
		if err != nil {
			t.Error(err)
		}
		_, err = discard.Write(testData)
		if err != nil {
			t.Error("cannot write after removing deadline")
		}
	})
}

func TestDiscardConn_Read(t *testing.T) {
	t.Run("read block then close", func(t *testing.T) {
		discard := Discard()
		done := make(chan struct{})
		go func() {
			_, _ = discard.Read(make([]byte, 1))
			done <- struct{}{}
		}()

		_ = discard.Close()
		select {
		case <-done:
			return
		case <-time.After(1 * time.Second):
			t.Error("Read did not unblock 1s after closing")
		}
	})

	t.Run("read block then timeout", func(t *testing.T) {
		discard := Discard()
		done := make(chan struct{})
		go func() {
			_, _ = discard.Read(make([]byte, 1))
			done <- struct{}{}
		}()

		_ = discard.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		select {
		case <-done:
			return
		case <-time.After(1 * time.Second):
			t.Error("Read did not unblock after deadline has passed")
		}
	})

	t.Run("racily setting deadline", func(t *testing.T) {
		discard := Discard()
		done := make(chan struct{})
		go func() {
			_, _ = discard.Read(make([]byte, 1))
			done <- struct{}{}
		}()

		for i := 5; i <= 105; i++ {
			go func() {
				_ = discard.SetDeadline(time.Now().Add(time.Duration(i) * time.Second))
			}()
		}

		// here we don't want read to return earlier than time.After(500ms)
		select {
		case <-done:
			t.Error("Read unblocked when deadline hasn't passed")
		case <-time.After(500 * time.Millisecond):
			return
		}
	})
}

func TestDiscardConn_Addrs(t *testing.T) {
	discard := Discard()
	if discard.LocalAddr() == nil {
		t.Error("LocalAddr shouldn't return null pointer")
	}
	if discard.RemoteAddr() == nil {
		t.Error("RemoteAddr shouldn't return null pointer")
	}
}
