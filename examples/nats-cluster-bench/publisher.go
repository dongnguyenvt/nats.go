package main

import (
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/bench"
)

type Publisher struct {
	numMsgs int
	msgSize int
	subject string
	donewg  *sync.WaitGroup
	nc      *nats.Conn
}

func NewPublisher(urls, subject string, numMsgs, msgSize int, donewg *sync.WaitGroup, opts ...nats.Option) (*Publisher, error) {
	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		return nil, err
	}
	return &Publisher{
		subject: subject,
		numMsgs: numMsgs,
		msgSize: msgSize,
		donewg:  donewg,
		nc:      nc,
	}, nil
}

func (p *Publisher) run() {
	var msg []byte
	if p.msgSize > 0 {
		msg = make([]byte, p.msgSize)
	}

	start := time.Now()

	for i := 0; i < p.numMsgs; i++ {
		p.nc.Publish(p.subject, msg)
	}
	p.nc.Flush()
	benchmark.AddPubSample(bench.NewSample(p.numMsgs, p.msgSize, start, time.Now(), p.nc))

	p.donewg.Done()
	p.nc.Close()
}
