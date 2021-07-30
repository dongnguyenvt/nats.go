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
	donewg  *sync.WaitGroup
	times   chan time.Time
	nc      *nats.Conn
}

func NewSubscriber(urls, subject string, numMsgs, msgSize int, donewg *sync.WaitGroup, opts ...nats.Option) (*Subscriber, error) {
	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		return nil, err
	}
	s := &Subscriber{
		subject: subject,
		numMsgs: numMsgs,
		msgSize: msgSize,
		donewg:  donewg,
		times:   make(chan time.Time, 2),
		nc:      nc,
	}
	return s, s.init()
}

func (s *Subscriber) init() error {
	received := 0
	sub, _ := s.nc.Subscribe(s.subject, func(msg *nats.Msg) {
		received++
		if received == 1 {
			s.times <- time.Now()
		}
		if received >= s.numMsgs {
			s.times <- time.Now()
		}
	})
	if err := sub.SetPendingLimits(-1, -1); err != nil {
		return err
	}
	return s.nc.Flush()
}

func (s *Subscriber) run() {
	start := <-s.times
	end := <-s.times
	benchmark.AddSubSample(bench.NewSample(s.numMsgs, s.msgSize, start, end, s.nc))
	s.nc.Close()
	s.donewg.Done()
}
