// Copyright 2015-2021 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/bench"
)

type Publisher struct {
	numMsgs int
	msgSize int
	subject string
	nc      *nats.Conn
}

func NewPublisher(urls, subject string, numMsgs, msgSize int, opts ...nats.Option) (*Publisher, error) {
	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		return nil, err
	}
	return &Publisher{
		subject: subject,
		numMsgs: numMsgs,
		msgSize: msgSize,
		nc:      nc,
	}, nil
}

func (p *Publisher) Run() *bench.Sample {
	var msg []byte
	if p.msgSize > 0 {
		msg = make([]byte, p.msgSize)
	}

	start := time.Now()

	for i := 0; i < p.numMsgs; i++ {
		// FIXME: ignore error for bench
		_ = p.nc.Publish(p.subject, msg)
	}
	// FIXME: ignore error for bench
	_ = p.nc.Flush()
	sample := bench.NewSample2(p.numMsgs, p.msgSize, start, time.Now(), p.nc.OutMsgs+p.nc.InMsgs, p.nc.OutBytes+p.nc.InBytes)
	p.nc.Close()
	return sample
}
