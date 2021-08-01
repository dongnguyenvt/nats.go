package main

import (
	"context"
	"errors"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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

func CreateContainer(image, name, port, networkname string, cmds []string) error {
	containers, err := ListContainer(name)
	if err != nil {
		return err
	}
	for _, c := range containers {
		if c.State == "running" {
			return nil
		}
	}
	if err = RemoveContainer(name); err != nil {
		return nil
	}
	ctx := context.Background()
	networkid, ok := GetNetworkID(networkname)
	if !ok {
		return errors.New(networkname + " is not valid network")
	}
	var portset nat.PortSet
	if len(port) > 0 {
		portset = nat.PortSet{nat.Port(port): {}}
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:        image,
			Cmd:          cmds,
			ExposedPorts: portset,
			Hostname:     name,
		}, nil, &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkname: {
					NetworkID: networkid,
				},
			},
		}, nil, name)
	if err != nil {
		return err
	}

	return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}

func ListContainer(name string) ([]types.Container, error) {
	ctx := context.Background()
	return cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", name)),
	})
}

func RemoveContainer(name string) error {
	ctx := context.Background()
	containers, err := ListContainer(name)
	if err != nil {
		return err
	}
	for _, c := range containers {
		if c.State == "running" {
			if err = cli.ContainerStop(ctx, c.ID, nil); err != nil {
				return err
			}
		}
		if err = cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
		}); err != nil {
			return err
		}
	}
	return nil
}
