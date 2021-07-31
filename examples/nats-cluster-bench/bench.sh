#!/usr/bin/env bash

set -x
set -e

NETWORK_NAME="natsbench"
CLUSTER_NAME="test-nats-cluster-bench"

# TODO configurable
NAT_SERVER_IMAGE="nats:latest"

# TODO arguments
CLUSTER_SIZE=3
NUM_PUBLISHERS=6
NUM_SUBSCRIBERS=6
MESSAGE_SIZE=1000

create_network()
{
  existing="$(docker network ls | grep "$NETWORK_NAME")"
  if [ -z "$existing" ]
  then
    docker network create "$NETWORK_NAME"
  fi
}

cleanup()
{
  containers=$(docker ps -a | grep "$NAT_SERVER_IMAGE" | awk '{print $1}')
  for container in $containers
  do
    docker stop "$container"
  done
}

start_cluster()
{
  # TODO: docker-compose
  docker run --rm -p 4222:4222 --name natserver1 --net $NETWORK_NAME -d "$NAT_SERVER_IMAGE" -p 4222 -cluster nats://natserver1:4248 --cluster_name "$CLUSTER_NAME"
  docker run --rm -p 5222:5222 --name natserver2 --net $NETWORK_NAME -d "$NAT_SERVER_IMAGE" -p 5222 -cluster nats://natserver2:5248 -routes nats://natserver1:4248 --cluster_name "$CLUSTER_NAME"
  docker run --rm -p 6222:6222 --name natserver3 --net $NETWORK_NAME -d "$NAT_SERVER_IMAGE" -p 6222 -cluster nats://natserver3:6248 -routes nats://natserver1:4248 --cluster_name "$CLUSTER_NAME"
}

start()
{
  # TODO: docker-compose
  docker run --rm -p 4222:4222 --name natserver1 --net $NETWORK_NAME -d "$NAT_SERVER_IMAGE" -p 4222
}

create_network
cleanup
#start_cluster
start
