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
}

func (p Publisher) run(nc *nats.Conn, startwg, donewg *sync.WaitGroup) {
	startwg.Done()

	var msg []byte
	if p.msgSize > 0 {
		msg = make([]byte, p.msgSize)
	}

	start := time.Now()

	for i := 0; i < p.numMsgs; i++ {
		nc.Publish(p.subject, msg)
	}
	nc.Flush()
	benchmark.AddPubSample(bench.NewSample(p.numMsgs, p.msgSize, start, time.Now(), nc))

	donewg.Done()
}
