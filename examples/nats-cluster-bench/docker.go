package main

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var cli *client.Client

func init() {
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
}

func GetNetworkID(s string) (string, bool) {
	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", s),
			filters.Arg("driver", "bridge"),
		),
	})
	if err != nil {
		return "", false
	}
	if len(networks) == 1 {
		return networks[0].ID, true
	}
	return "", false
}

func CreateNetwork(s string) error {
	if _, ok := GetNetworkID(s); ok {
		return nil
	}
	_, err := cli.NetworkCreate(context.Background(), s, types.NetworkCreate{
		Driver: "bridge",
	})
	return err
}

func DeleteNetwork(s string) error {
	if id, ok := GetNetworkID(s); ok {
		return cli.NetworkRemove(context.Background(), id)
	}
	return nil
}
