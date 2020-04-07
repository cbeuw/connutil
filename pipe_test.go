package connutil

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var lengths = []struct {
	length int
}{
	{1},
	{2},
	{3},
	{128},
	{192},
	{256},
	{65536},
	{1 << 20},
	{1 << 28},
}

func TestPipeConn_Write(t *testing.T) {
	testData := make([]byte, 128)
	t.Run("write after close", func(t *testing.T) {
		_, w := AsyncPipe()
		err := w.Close()
		if err != nil {
			t.Error(err)
		}
		_, err = w.Write(testData)
		if err != io.ErrClosedPipe {
			t.Errorf("expecting %v, got %v", io.ErrClosedPipe, err)
		}
	})
	t.Run("write after deadline passed", func(t *testing.T) {
		_, w := AsyncPipe()
		err := w.SetWriteDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, err = w.Write(testData)
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
	t.Run("recover from write deadline", func(t *testing.T) {
		_, w := AsyncPipe()
		err := w.SetWriteDeadline(time.Now().Add(-1 * time.Second))
		_, err = w.Write(testData)
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
		err = w.SetWriteDeadline(time.Time{})
		if err != nil {
			t.Error(err)
		}
		_, err = w.Write(testData)
		if err != nil {
			t.Error("cannot write after removing deadline")
		}
	})
}

func TestPipeConn_Read(t *testing.T) {
	makeBuf := func() []byte {
		return make([]byte, 128)
	}
	t.Run("read after close", func(t *testing.T) {
		r, _ := AsyncPipe()
		err := r.Close()
		if err != nil {
			t.Error(err)
		}
		_, err = r.Read(makeBuf())
		if err != io.ErrClosedPipe {
			t.Errorf("expecting %v, got %v", io.ErrClosedPipe, err)
		}
	})
	t.Run("read after deadline passed", func(t *testing.T) {
		r, _ := AsyncPipe()
		err := r.SetReadDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, err = r.Read(makeBuf())
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
	t.Run("recover from read deadline", func(t *testing.T) {
		r, w := AsyncPipe()
		err := r.SetReadDeadline(time.Now().Add(-1 * time.Second))
		_, err = r.Read(makeBuf())
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
		err = r.SetReadDeadline(time.Time{})
		if err != nil {
			t.Error(err)
		}
		_, _ = w.Write([]byte{1, 2, 3})
		_, err = r.Read(makeBuf())
		if err != nil {
			t.Error("cannot read after removing deadline")
		}
	})
}

func TestPipeConn_ReadWrite(t *testing.T) {
	testData := make([]byte, 128)
	t.Run("simple read write", func(t *testing.T) {
		r, w := AsyncPipe()
		_, err := w.Write(testData)
		if err != nil {
			t.Error(err)
		}

		receiveBuf := make([]byte, len(testData))
		_, err = io.ReadFull(r, receiveBuf)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("read write after deadline passed", func(t *testing.T) {
		c, _ := AsyncPipe()
		err := c.SetDeadline(time.Now().Add(-1 * time.Second))
		if err != nil {
			t.Error(err)
		}
		_, err = c.Write(testData)
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}

		_, err = c.Read(make([]byte, len(testData)))
		if err != ErrTimeout {
			t.Errorf("expecting %v, got %v", ErrTimeout, err)
		}
	})
}

func TestPipeConn_concurrentPipe(t *testing.T) {
	for _, tt := range lengths {
		t.Run(fmt.Sprint(tt.length), func(t *testing.T) {
			w, r := AsyncPipe()
			content := make([]byte, tt.length)
			contentLen := len(content)

			go func() {
				readBuf := make([]byte, contentLen)
				_, err := io.ReadFull(r, readBuf)
				if err != nil {
					t.Errorf("failed to read full: %v", err)
				}
			}()

			n, err := w.Write(content)
			if err != nil {
				t.Errorf("failed to write %v", err)
			}
			if n != len(content) {
				t.Errorf("expected to write %v but written %v", len(content), n)
			}
		})
	}
}

func TestPipeConn_concurrentPipe_tricklingWrite(t *testing.T) {
	for _, tt := range lengths {
		t.Run(fmt.Sprint(tt.length), func(t *testing.T) {
			w, r := AsyncPipe()
			content := make([]byte, tt.length)
			contentLen := tt.length

			go func() {
				readBuf := make([]byte, contentLen)
				_, err := io.ReadFull(r, readBuf)
				if err != nil {
					t.Errorf("failed to read full: %v", err)
				}
			}()

			remaining := contentLen
			for remaining != 0 {
				toWrite := rand.Intn(remaining) + 1
				_, err := w.Write(content[contentLen-remaining : contentLen-remaining+toWrite])
				if err != nil {
					t.Errorf("failed to write %v", err)
				}
				remaining -= toWrite
				time.Sleep(1 * time.Millisecond)
			}
		})
	}
}

func TestAsyncPipe_Duplex(t *testing.T) {
	test := func(dataLen int) {
		pWriteData := make([]byte, dataLen)
		qWriteData := make([]byte, dataLen)
		rand.Read(pWriteData)
		rand.Read(qWriteData)

		pReadBuf := make([]byte, dataLen)
		qReadBuf := make([]byte, dataLen)
		p, q := AsyncPipe()
		var wg sync.WaitGroup
		go func() {
			_, err := p.Write(pWriteData)
			if err != nil {
				t.Fatal(err)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := io.ReadFull(q, qReadBuf)
			if err != nil {
				t.Fatal(err)
			}
		}()
		go func() {
			_, err := q.Write(qWriteData)
			if err != nil {
				t.Fatal(err)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := io.ReadFull(p, pReadBuf)
			if err != nil {
				t.Fatal(err)
			}
		}()

		wg.Wait()
		if !bytes.Equal(qWriteData, pReadBuf) {
			t.Error("data written to q not correctly read by p")
		}
		if !bytes.Equal(pWriteData, qReadBuf) {
			t.Error("data written to p not correctly read by q")
		}
	}

	for i := 0; i < 20; i += 2 {
		dataLen := 1 << i
		t.Run("", func(t *testing.T) {
			t.Parallel()
			test(dataLen)
		})
	}
}

func TestLimitedAsyncPipe(t *testing.T) {
	testData := make([]byte, 1<<16)
	rand.Read(testData)
	t.Run("large write", func(t *testing.T) {
		w, r := LimitedAsyncPipe(1)
		go func() {
			contentLen := len(testData)
			remaining := contentLen
			for remaining != 0 {
				toWrite := rand.Intn(remaining) + 1
				_, _ = w.Write(testData[contentLen-remaining : contentLen-remaining+toWrite])
				remaining -= toWrite
			}
		}()
		readBuf := make([]byte, len(testData))
		_, err := io.ReadFull(r, readBuf)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(testData, readBuf) {
			t.Error("data not correctly read")
		}
	})
}
