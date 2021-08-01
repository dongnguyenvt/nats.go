package main

import (
	"encoding/json"
	"os"
	"testing"
)

func newTestSuite() TestSuite {
	return TestSuite{
		CleanUpAfterEachTest: false,
		Tls:                  false,
		UserCreds:            "",
		NkeyFile:             "",
		TestCases: []Instance{
			{
				ClusterSize:  1,
				NumPubs:      1,
				NumSubs:      0,
				NumMsgs:      6000,
				MsgSize:      32,
				RandomScheme: "cryptorand",
			},
			{
				ClusterSize:  1,
				NumPubs:      1,
				NumSubs:      2,
				NumMsgs:      6000,
				MsgSize:      32,
				RandomScheme: "cryptorand",
			},
			{
				ClusterSize:  2,
				NumPubs:      2,
				NumSubs:      4,
				NumMsgs:      6000,
				MsgSize:      32,
				RandomScheme: "cryptorand",
			},
			{
				ClusterSize:  3,
				NumPubs:      3,
				NumSubs:      6,
				NumMsgs:      6000,
				MsgSize:      32,
				RandomScheme: "cryptorand",
			},
		},
	}
}

func TestCreateTestSuite(t *testing.T) {
	ts := newTestSuite()
	if err := ts.Validate(); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(ts)
	t.Log(string(data), err)
}

func TestLoadTestSuite(t *testing.T) {
	f, err := os.Open("test_suite.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	var ts TestSuite
	if err = decoder.Decode(&ts); err != nil {
		t.Fatal(err)
	}
	if err = ts.Validate(); err != nil {
		t.Fatal(err)
	}
	t.Log(ts)
	//if err = ts.Setup(); err != nil {
	//	t.Fatal(err)
	//}
	//if err = ts.Run(); err != nil {
	//	t.Fatal(err)
	//}
	//if err = ts.Cleanup(); err != nil {
	//	t.Fatal(err)
	//}
}
