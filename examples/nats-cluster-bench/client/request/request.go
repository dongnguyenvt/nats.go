package request

import (
	"encoding/json"
	"net/http"
)

type ClientMode int

const (
	Unknown ClientMode = iota
	Publisher
	Subscriber
)

// Init test request
// TODO: nats options
type Init struct {
	Mode           ClientMode `json:"mode"`
	NatsServerUrls []string   `json:"nats_server_urls"`
	Subject        string     `json:"subject"`
	NumMsgs        int        `json:"num_msgs"`
	MsgSize        int        `json:"msg_size"`
}

func ParseInitReq(req *http.Request) (Init, error) {
	decoder := json.NewDecoder(req.Body)
	var data Init
	err := decoder.Decode(&data)
	return data, err
}
