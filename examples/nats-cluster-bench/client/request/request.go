package request

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
)

type ClientMode int
type RandomScheme int

const (
	Unknown ClientMode = iota
	Publisher
	Subscriber
)

const (
	None RandomScheme = iota
	MathRand
	CryptoRand
)

// Options are nats options
// Note: UserCreds and NkeyFile paths should be available on client image
type Options struct {
	Name      string `json:"name"`
	UserCreds string `json:"user_creds"`
	NkeyFile  string `json:"nkey_file"`
	Tls       bool   `json:"tls"`
}

func (o Options) NatsOptions() (opts []nats.Option) {
	if len(o.Name) > 0 {
		opts = append(opts, nats.Name(o.Name))
	}
	// Use UserCredentials
	if o.UserCreds != "" {
		opts = append(opts, nats.UserCredentials(o.UserCreds))
	}

	// Use Nkey authentication.
	if o.NkeyFile != "" {
		opt, err := nats.NkeyOptionFromSeed(o.NkeyFile)
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts, opt)
	}

	// Use TLS specified
	if o.Tls {
		opts = append(opts, nats.Secure(nil))
	}

	return
}

type ClientConfig struct {
	Mode ClientMode `json:"mode"`
	// for publisher
	RandomScheme `json:"random_scheme"`
}

// Init test request
type Init struct {
	Config         ClientConfig `json:"config"`
	NatsServerUrls []string     `json:"nats_server_urls"`
	Subject        string       `json:"subject"`
	NumMsgs        int          `json:"num_msgs"`
	MsgSize        int          `json:"msg_size"`
	Options        Options      `json:"options"`
}

func ParseInitReq(req *http.Request) (Init, error) {
	decoder := json.NewDecoder(req.Body)
	var data Init
	err := decoder.Decode(&data)
	return data, err
}

func (scheme RandomScheme) String() string {
	switch scheme {
	case MathRand:
		return "mathrand"
	case CryptoRand:
		return "cryptorand"
	case None:
		fallthrough
	default:
		return "none"
	}
}

func RandomSchemeFromString(s string) RandomScheme {
	switch s {
	case "mathrand":
		return MathRand
	case "cryptorand":
		return CryptoRand
	default:
		return None
	}
}
