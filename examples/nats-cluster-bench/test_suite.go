package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nats-io/nats.go/bench"
	"github.com/nats-io/nats.go/examples/nats-cluster-bench/client/request"
)

const (
	clusterName      = "test-nats-cluster-bench"
	networkName      = "natsbench"
	serverImage      = "nats:latest"
	clientImage      = "local/nats-client:latest"
	serverNamePrefix = "nats-server-"
	clientNamePrefix = "nats-client-"
)

// Instance is a test instance configuration
type Instance struct {
	urls         []string
	benchmark    *bench.Benchmark
	ClusterSize  int    `json:"cluster_size"`
	NumPubs      int    `json:"num_pubs"`
	NumSubs      int    `json:"num_subs"`
	NumMsgs      int    `json:"num_msgs"`
	MsgSize      int    `json:"msg_size"`
	RandomScheme string `json:"random_scheme"`
	rs           request.RandomScheme
}

// TestSuite is test orchestrator
// for each test instance: setup, run test and collect stats
// render final output
type TestSuite struct {
	DoCleanUp            bool       `json:"clean_up"`
	CleanUpAfterEachTest bool       `json:"clean_up_after_each_test"`
	Tls                  bool       `json:"tls"`
	UserCreds            string     `json:"user_creds"`
	NkeyFile             string     `json:"nkey_file"`
	TestCases            []Instance `json:"test_cases"`
}

func (t *Instance) Validate() error {
	if t.ClusterSize <= 0 {
		return errors.New("cluster size must be positive")
	}
	if t.NumSubs < 0 {
		return errors.New("num subscriber is negative")
	}
	if t.NumPubs < 0 {
		return errors.New("num publisher is negative")
	}
	if t.NumMsgs <= 0 {
		return errors.New("num msg must be positive")
	}
	if t.MsgSize <= 0 {
		return errors.New("msg size must be positive")
	}
	if (t.NumSubs%t.ClusterSize)+(t.NumPubs%t.ClusterSize)+(t.NumMsgs%t.ClusterSize) > 0 {
		return errors.New("num publisher/subscriber and message size must be multiplier of cluster size")
	}
	t.rs = request.RandomSchemeFromString(t.RandomScheme)
	if t.rs == request.None && t.RandomScheme != "none" {
		return errors.New("invalid random scheme")
	}
	return nil
}

func (t *Instance) Setup() error {
	t.benchmark = bench.NewBenchmark("NATS", t.NumSubs, t.NumPubs)
	var seedServerUrl string
	for i := 0; i < t.ClusterSize; i++ {
		var (
			serverName = serverNamePrefix + strconv.Itoa(i)
			routes     = fmt.Sprintf("nats://%s:4248", serverName)
			cmds       = []string{
				"-p",
				"4222",
				"-cluster",
				routes,
				"--cluster_name",
				clusterName,
			}
		)
		if i == 0 {
			seedServerUrl = routes
		} else {
			cmds = append(cmds, "-routes")
			cmds = append(cmds, seedServerUrl)
		}
		if err := CreateContainer(serverImage, serverName, "", networkName, cmds); err != nil {
			return err
		}
	}
	// TODO: configurable
	// use for signaling clients
	const startingPort = 8080
	for i := 0; i < t.NumSubs+t.NumPubs; i++ {
		var cmds = []string{
			"-p",
			strconv.Itoa(startingPort + i),
		}
		if err := CreateContainer(clientImage, clientNamePrefix+strconv.Itoa(i), strconv.Itoa(startingPort+i), networkName, cmds); err != nil {
			return err
		}
	}
	return nil
}

func (t *Instance) Run() error {
	return nil
}

func (t *Instance) Cleanup() error {
	return nil
}

func (t *TestSuite) Setup() error {
	if err := CreateNetwork(networkName); err != nil {
		return err
	}
	return nil
}

func (t *TestSuite) Run() error {
	for _, tc := range t.TestCases {
		if err := tc.Setup(); err != nil {
			return err
		}
		if err := tc.Run(); err != nil {
			return err
		}
		if err := tc.Cleanup(); err != nil {
			return err
		}
		if t.DoCleanUp && t.CleanUpAfterEachTest {
			if err := t.Cleanup(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *TestSuite) Cleanup() error {
	if !t.DoCleanUp {
		return nil
	}
	container, err := ListContainer("")
	if err != nil {
		return err
	}
	for _, c := range container {
		for _, name := range c.Names {
			if strings.Contains(name, clientNamePrefix) {
				if err = RemoveContainer(name); err != nil {
					return err
				}
				break
			}
			if strings.Contains(name, serverNamePrefix) {
				if err = RemoveContainer(name); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func (t *TestSuite) Validate() error {
	if t.UserCreds != "" && t.NkeyFile != "" {
		return errors.New("specify -nkey or -creds")
	}
	for _, tc := range t.TestCases {
		if err := tc.Validate(); err != nil {
			return err
		}
	}
	return nil
}
