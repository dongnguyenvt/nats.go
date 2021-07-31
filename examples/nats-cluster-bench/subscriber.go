package main

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/bench"
)

type Subscriber struct {
	numMsgs int
	msgSize int
	subject string
	times   chan time.Time
	nc      *nats.Conn
}

func NewSubscriber(urls, subject string, numMsgs, msgSize int, opts ...nats.Option) (*Subscriber, error) {
	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		return nil, err
	}
	s := &Subscriber{
		subject: subject,
		numMsgs: numMsgs,
		msgSize: msgSize,
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

func (s *Subscriber) run() *bench.Sample {
	start := <-s.times
	end := <-s.times
	sample := bench.NewSample2(s.numMsgs, s.msgSize, start, end, s.nc.OutMsgs+s.nc.InMsgs, s.nc.OutBytes+s.nc.InBytes)
	s.nc.Close()
	return sample
}
