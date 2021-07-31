package request

import (
	"encoding/json"
	"net/http"

	"github.com/nats-io/nats.go"
)

type ClientMode int

const (
	Unknown ClientMode = iota
	Publisher
	Subscriber
)

// Options are nats options
type Options struct {
	Name string `json:"name"`
}

func (o Options) NatsOptions() (opts []nats.Option) {
	// TODO: more options
	if len(o.Name) > 0 {
		opts = append(opts, nats.Name(o.Name))
	}
	return
}

// Init test request
type Init struct {
	Mode           ClientMode `json:"mode"`
	NatsServerUrls []string   `json:"nats_server_urls"`
	Subject        string     `json:"subject"`
	NumMsgs        int        `json:"num_msgs"`
	MsgSize        int        `json:"msg_size"`
	Options        Options    `json:"options"`
}

func ParseInitReq(req *http.Request) (Init, error) {
	decoder := json.NewDecoder(req.Body)
	var data Init
	err := decoder.Decode(&data)
	return data, err
}
