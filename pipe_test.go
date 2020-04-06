package connutil

import (
	"fmt"
	"io"
	"math/rand"
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

func TestAsyncPipe_concurrentPipe(t *testing.T) {
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

func TestAsyncPipe_concurrentPipe_tricklingWrite(t *testing.T) {
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
