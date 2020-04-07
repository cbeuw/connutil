package connutil

import (
	"io"
	"math/rand"
	"testing"
	"time"
)

func TestBabelConn_Write(t *testing.T) {
	testData := make([]byte, 128)
	t.Run("simple write", func(t *testing.T) {
		_, err := Discard().Write(testData)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("write after close", func(t *testing.T) {
		babel := Babel(rand.New(rand.NewSource(0)))
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
		babel := Babel(rand.New(rand.NewSource(0)))
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
		babel := Babel(rand.New(rand.NewSource(0)))
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
