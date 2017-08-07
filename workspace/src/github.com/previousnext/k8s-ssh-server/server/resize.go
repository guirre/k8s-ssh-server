package main

import (
	"github.com/gliderlabs/ssh"
	"k8s.io/kubernetes/pkg/util/term"
)

// SizeQueue stores window resize events.
type SizeQueue struct {
	resizeChan chan term.Size
}

// NewResizeQueue returns a size queue for storing window resize events.
func NewResizeQueue(sess ssh.Session) *SizeQueue {
	queue := &SizeQueue{
		resizeChan: make(chan term.Size, 1),
	}

	_, winCh, _ := sess.Pty()
	go func() {
		for win := range winCh {
			queue.resizeChan <- term.Size{
				Height: uint16(win.Height),
				Width:  uint16(win.Width),
			}
		}
	}()

	return queue
}

// Next returns the next window resize event.
func (s *SizeQueue) Next() *term.Size {
	size, ok := <-s.resizeChan
	if !ok {
		return nil
	}
	return &size
}
