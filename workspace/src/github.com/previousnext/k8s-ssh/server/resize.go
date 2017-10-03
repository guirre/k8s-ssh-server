package main

import (
	"github.com/gliderlabs/ssh"
	"k8s.io/client-go/tools/remotecommand"
)

// SizeQueue stores window resize events.
type SizeQueue struct {
	resizeChan chan remotecommand.TerminalSize
}

// NewResizeQueue returns a size queue for storing window resize events.
func NewResizeQueue(sess ssh.Session) *SizeQueue {
	queue := &SizeQueue{
		resizeChan: make(chan remotecommand.TerminalSize, 1),
	}

	_, winCh, _ := sess.Pty()
	go func() {
		for win := range winCh {
			queue.resizeChan <- remotecommand.TerminalSize{
				Height: uint16(win.Height),
				Width:  uint16(win.Width),
			}
		}
	}()

	return queue
}

// Next returns the next window resize event.
func (s *SizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-s.resizeChan
	if !ok {
		return nil
	}
	return &size
}
