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
