package main

import "testing"

func TestCreateNetwork(t *testing.T) {
	if err := CreateNetwork("test"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteNetwork(t *testing.T) {
	if err := CreateNetwork("test"); err != nil {
		t.Fatal(err)
	}
	if err := DeleteNetwork("test"); err != nil {
		t.Fatal(err)
	}
	if err := DeleteNetwork("test"); err != nil {
		t.Fatal(err)
	}
}

func TestCreateContainer(t *testing.T) {
	if err := CreateNetwork("testnetwork"); err != nil {
		t.Fatal(err)
	}
	if err := CreateContainer("nats:latest", "test-natsserver", "", "testnetwork", nil); err != nil {
		t.Fatal(err)
	}
	if err := RemoveContainer("test-natsserver"); err != nil {
		t.Fatal(err)
	}
}
