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

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/bench"
	"github.com/nats-io/nats.go/examples/nats-cluster-bench/client/request"
)

// Some sane defaults
const (
	DefaultNumMsgs     = 100000
	DefaultNumPubs     = 1
	DefaultNumSubs     = 0
	DefaultMessageSize = 128
)

func usage() {
	log.Printf("Usage: nats-bench [-s server (%s)] [--tls] [-np NUM_PUBLISHERS] [-ns NUM_SUBSCRIBERS] [-n NUM_MSGS] [-ms MESSAGE_SIZE] [-csv csvfile] [-creds file] [-nkey file] <subject>\n", nats.DefaultURL)
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}

var benchmark *bench.Benchmark

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	//var tls = flag.Bool("tls", false, "Use TLS Secure Connection")
	var numPubs = flag.Int("np", DefaultNumPubs, "Number of Concurrent Publishers")
	var numSubs = flag.Int("ns", DefaultNumSubs, "Number of Concurrent Subscribers")
	var numMsgs = flag.Int("n", DefaultNumMsgs, "Number of Messages to Publish")
	var msgSize = flag.Int("ms", DefaultMessageSize, "Size of the message.")
	var csvFile = flag.String("csv", "", "Save bench data to csv file")
	//var userCreds = flag.String("creds", "", "User Credentials File")
	//var nkeyFile = flag.String("nkey", "", "NKey Seed File")
	var showHelp = flag.Bool("h", false, "Show help message")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		showUsageAndExit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		showUsageAndExit(1)
	}

	if *numMsgs <= 0 {
		log.Fatal("Number of messages should be greater than zero.")
	}

	// Connect Options.
	//opts := []nats.Option{nats.Name("NATS Benchmark")}
	//
	//if *userCreds != "" && *nkeyFile != "" {
	//	log.Fatal("specify -seed or -creds")
	//}
	//
	//// Use UserCredentials
	//if *userCreds != "" {
	//	opts = append(opts, nats.UserCredentials(*userCreds))
	//}
	//
	//// Use Nkey authentication.
	//if *nkeyFile != "" {
	//	opt, err := nats.NkeyOptionFromSeed(*nkeyFile)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	opts = append(opts, opt)
	//}
	//
	//// Use TLS specified
	//if *tls {
	//	opts = append(opts, nats.Secure(nil))
	//}

	var subj = args[0]

	var port = 8080
	// Run Subscribers first
	// This order is important for getting reliable benchmark result
	// if publishers are running first, we are losing messages to void
	// also waiting for Subscribers are done serves as synchronization method to signal test end.
	// Note that we don't need to explicitly synchronize waiting for publishers.
	for i := 0; i < *numSubs; i++ {
		initReq := request.Init{
			Config: request.ClientConfig{
				Mode: request.Subscriber,
			},
			NatsServerUrls: strings.Split(*urls, ","),
			Subject:        subj,
			NumMsgs:        *numMsgs,
			MsgSize:        *msgSize,
			Options: request.Options{
				Name: "NATS Benchmark",
			},
		}
		data, err := json.Marshal(initReq)
		if err != nil {
			log.Fatalf("Init Subscriber failed: %v", err)
		}
		res, err := http.Post("http://localhost:"+strconv.Itoa(port)+"/init", "application/json", bytes.NewReader(data))
		if err != nil {
			log.Fatalf("Init Subscriber failed: %v", err)
		}
		response, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Init Subscriber failed: %v", err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			log.Fatalf("Init Subscriber failed: %s", string(response))
		}
		port++
	}

	// Now Publishers
	pubCounts := bench.MsgsPerClient(*numMsgs, *numPubs)
	for i := 0; i < *numPubs; i++ {
		initReq := request.Init{
			Config: request.ClientConfig{
				Mode:         request.Publisher,
				RandomScheme: request.MathRand, // TODO: configurable
			},
			NatsServerUrls: strings.Split(*urls, ","),
			Subject:        subj,
			NumMsgs:        pubCounts[i],
			MsgSize:        *msgSize,
			Options: request.Options{
				Name: "NATS Benchmark",
			},
		}
		data, err := json.Marshal(initReq)
		if err != nil {
			log.Fatalf("Init Publisher failed: %v", err)
		}
		res, err := http.Post("http://localhost:"+strconv.Itoa(port)+"/init", "application/json", bytes.NewReader(data))
		if err != nil {
			log.Fatalf("Init Publisher failed: %v", err)
		}
		response, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Init Publisher failed: %v", err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			log.Fatalf("Init Publisher failed: %s", string(response))
		}
		port++
	}

	log.Printf("Starting benchmark [msgs=%d, msgsize=%d, pubs=%d, subs=%d]\n", *numMsgs, *msgSize, *numPubs, *numSubs)

	var wg sync.WaitGroup
	wg.Add(*numSubs + *numPubs)
	benchmark = bench.NewBenchmark("NATS", *numSubs, *numPubs)
	port = 8080
	for i := 0; i < *numSubs+*numPubs; i++ {
		go func(i int) {
			defer wg.Done()
			res, err := http.Get("http://localhost:" + strconv.Itoa(port+i) + "/start")
			if err != nil {
				log.Fatalf("test run failed: %v", err)
			}
			response, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatalf("test run failed: %v", err)
			}
			_ = res.Body.Close()
			if res.StatusCode != http.StatusOK {
				log.Fatalf("test run failed: %s", string(response))
			}
			var sample bench.Sample
			err = json.Unmarshal(response, &sample)
			if err != nil {
				log.Fatalf("test run failed: %v", err)
			}
			if i < *numSubs {
				benchmark.AddSubSample(&sample)
			} else {
				benchmark.AddPubSample(&sample)
			}
		}(i)
	}
	wg.Wait()
	benchmark.Close()

	fmt.Print(benchmark.Report())

	if len(*csvFile) > 0 {
		csv := benchmark.CSV()
		ioutil.WriteFile(*csvFile, []byte(csv), 0644)
		fmt.Printf("Saved metric data in csv file %s\n", *csvFile)
	}
}
