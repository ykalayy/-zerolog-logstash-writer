package logstash

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type logstashWriter struct {
	// Writer options
	option *Option
	// Runtime log buffer
	logBuffer chan []byte
	// Interrupt function of writer
	context.CancelFunc
	// Tcp connection to Logstash
	conn net.Conn
	// Writer sync. FIFO security
	// TODO: What about semaphore
	wLock sync.WaitGroup
}

// NewAsyncWriter Returns an async recoverable logstash writer
func NewAsyncWriter(ops *Option) (*logstashWriter, context.CancelFunc) {

	ops.init()

	logstashWr := &logstashWriter{
		option:    ops,
		logBuffer: make(chan []byte, ops.BufferSize),
		wLock:     sync.WaitGroup{},
	}
	ctx, cancelFunc := context.WithCancel(context.Background())

	// Single I/O writer to logstash socket
	go logstashWr.writerTask(ctx)

	return logstashWr, cancelFunc
}

// Logstash async I/O routine
func (w *logstashWriter) writerTask(ctx context.Context) {

	for {
		select {
		case logData := <-w.logBuffer:
			if w.conn != nil {
				_, err := w.conn.Write(logData)
				if err != nil {
					log.Println("Connection is down")
					// Trying to re-establish the connection
					w.conn, _ = setupDial(0, w.option)
				}
			} else {
				// First attempt for logstash connection, Initially It only tries once
				// TODO: Lazy connection initialization can be a problem or not?
				w.conn, _ = setupDial(0, w.option)
			}
		case <-ctx.Done():
			log.Println("Writing to the Logstash is ended")
		}
	}
}

func (w *logstashWriter) Write(p []byte) (n int, err error) {

	// It is only for secure FIFO operation When the buffer is full
	w.wLock.Wait()

	select {
	case w.logBuffer <- p:
	default:
		// Buffer is full. Lock write
		w.wLock.Add(1)
		defer w.wLock.Done()
		// FIFO, Pop first item from the buffer
		select {
		case <-w.logBuffer:
		}
		// Write the latest log
		// It is still non-blocking to eliminate deadlock
		select {
		case w.logBuffer <- p:
		}
	}

	return len(p), nil // Assume write is success
}

// Retries until Dial
func setupDial(attemptCount int, ops *Option) (conn net.Conn, err error) {

	conn, err = net.Dial("tcp", ops.Addr)
	if err != nil {
		if attemptCount < ops.RetryCount {
			log.Printf("Error: Coundn't open connection to logstash... Trying again. AttemptCount: %v \n", attemptCount)
			// Sleep until DialTimeout
			time.Sleep(ops.DialTimeout)
			return setupDial(attemptCount+1, ops)
		} else {
			return nil, errors.New("connection establish is failed")
		}
	}

	return conn, err
}
