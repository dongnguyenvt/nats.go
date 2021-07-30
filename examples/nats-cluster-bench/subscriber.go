package main

import (
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/bench"
)

type Subscriber struct {
	numMsgs int
	msgSize int
	subject string
}

func (s Subscriber) run(nc *nats.Conn, startwg, donewg *sync.WaitGroup) {
	received := 0
	ch := make(chan time.Time, 2)
	sub, _ := nc.Subscribe(s.subject, func(msg *nats.Msg) {
		received++
		if received == 1 {
			ch <- time.Now()
		}
		if received >= s.numMsgs {
			ch <- time.Now()
		}
	})
	sub.SetPendingLimits(-1, -1)
	nc.Flush()
	startwg.Done()

	start := <-ch
	end := <-ch
	benchmark.AddSubSample(bench.NewSample(s.numMsgs, s.msgSize, start, end, nc))
	nc.Close()
	donewg.Done()
}
