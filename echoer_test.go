package connutil

import (
	"io"
	"math/rand"
	"testing"
)

func TestEchoer(t *testing.T) {
	e := Echoer()
	testData := make([]byte, 128)
	rand.Read(testData)
	_, err := e.Write(testData)
	if err != nil {
		t.Error(err)
	}
	recvBuf := make([]byte, len(testData))
	_, err = io.ReadFull(e, recvBuf)
	if err != nil {
		t.Error(err)
	}

	_ = e.Close()
	if _, err = e.Read(recvBuf); err != io.ErrClosedPipe {
		t.Errorf("expecting error %v, got %v", io.ErrClosedPipe, err)
	}
}
