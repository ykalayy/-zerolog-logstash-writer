package logstash

import "time"

const (
	defaultBufferSize  = 1024
	defaultRetryCount  = 5
	defaultDialTimeout = time.Second * 5
)

type Option struct {

	// Logstash Host:port
	Addr string
	// Log buffer size. Default:1024
	BufferSize int
	// DialTimeout Retry delay. Default: 5 sec
	DialTimeout time.Duration
	// RetryCount Max Reconnect Retry count for each log attempt
	RetryCount int
}

func (o *Option) init() {

	if o.Addr == "" {
		o.Addr = "localhost:5044"
	}

	if o.DialTimeout == 0 {
		o.DialTimeout = defaultDialTimeout
	}

	if o.BufferSize == 0 {
		o.BufferSize = defaultBufferSize
	}

	if o.RetryCount == 0 {
		o.RetryCount = defaultRetryCount
	}
}
