package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/nats-io/nats.go/examples/nats-cluster-bench/client"
	"github.com/nats-io/nats.go/examples/nats-cluster-bench/client/request"
)

func usage() {
	log.Println("Usage: nats-bench-Client -p port")
	flag.PrintDefaults()
}

func main() {
	var port = flag.String("p", "8080", "The nats bench Client listening port")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	c := &Client{}
	mux := http.DefaultServeMux
	mux.HandleFunc("/init", c.initBenchHandler)
	mux.HandleFunc("/start", c.startBenchHandler)

	log.Printf("nats-bench-Client start listening on port: %s\n", *port)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("nats-Client start http failed: %v", err)
	}
}

// FIXME: is there more intuitive way to handle test-bench session and reduce API calls?
// for now it requires test-orchestrator to init test session and then signal test run
func (c *Client) initBenchHandler(w http.ResponseWriter, r *http.Request) {
	if err := c.reset(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := request.ParseInitReq(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch data.Mode {
	case request.Publisher:
		c.pub, err = client.NewPublisher(
			strings.Join(data.NatsServerUrls, ","),
			data.Subject,
			data.NumMsgs,
			data.MsgSize,
			data.Options.NatsOptions()...,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case request.Subscriber:
		c.sub, err = client.NewSubscriber(
			strings.Join(data.NatsServerUrls, ","),
			data.Subject,
			data.NumMsgs,
			data.MsgSize,
			data.Options.NatsOptions()...,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "invalid client mode", http.StatusBadRequest)
		return
	}
	c.mode = data.Mode
	c.initialized = true
	c.run = false
}

func (c *Client) startBenchHandler(w http.ResponseWriter, _ *http.Request) {
	if !c.isInitialized() {
		http.Error(w, "test session is not initialized", http.StatusBadRequest)
		return
	}
	defer func() {
		c.run = false
	}()
	switch c.mode {
	case request.Publisher:
		if c.pub == nil {
			http.Error(w, "publisher not init", http.StatusInternalServerError)
			return
		}
		c.run = true
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(c.pub.Run())
	case request.Subscriber:
		if c.sub == nil {
			http.Error(w, "subscriber not init", http.StatusInternalServerError)
			return
		}
		c.run = true
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(c.sub.Run())
	default:
		http.Error(w, "something wrong", http.StatusInternalServerError)
		return
	}
}

func (c *Client) isInitialized() bool {
	return c.initialized && c.mode != request.Unknown
}

func (c *Client) isRunning() bool {
	return c.run
}

func (c *Client) reset() error {
	if c.isRunning() {
		return errors.New("test session already run")
	}
	c.mode = request.Unknown
	c.initialized = false
	c.run = false
	return nil
}

type Client struct {
	mode        request.ClientMode
	initialized bool
	run         bool
	pub         *client.Publisher
	sub         *client.Subscriber
}
