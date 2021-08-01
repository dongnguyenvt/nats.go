#!/bin/bash

# Copyright 2015-2021 The NATS Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -x
set -e

NETWORK_NAME="natsbench"
CLUSTER_NAME="test-nats-cluster-bench"

# TODO configurable
NAT_SERVER_IMAGE="nats:latest"
NAT_CLIENT_IMAGE="local/nats-client:latest"

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
  containers=$(docker ps -a | grep nats-server | awk '{print $1}')
  for container in $containers
  do
    docker stop "$container"
  done
  containers=$(docker ps -a | grep nats-client | awk '{print $1}')
  for container in $containers
  do
    docker stop "$container"
  done
}

start_cluster()
{
  # TODO: docker-compose
  docker run --rm -p 4222:4222 --name nats-server-1 --net $NETWORK_NAME -d "$NAT_SERVER_IMAGE" -p 4222 -cluster nats://nats-server-1:4248 --cluster_name "$CLUSTER_NAME"
  docker run --rm -p 5222:5222 --name nats-server-2 --net $NETWORK_NAME -d "$NAT_SERVER_IMAGE" -p 5222 -cluster nats://nats-server-2:5248 -routes nats://nats-server-1:4248 --cluster_name "$CLUSTER_NAME"
  docker run --rm -p 6222:6222 --name nats-server-3 --net $NETWORK_NAME -d "$NAT_SERVER_IMAGE" -p 6222 -cluster nats://nats-server-3:6248 -routes nats://nats-server-1:4248 --cluster_name "$CLUSTER_NAME"
}

start()
{
  # TODO: docker-compose
  docker run --rm -p 4222:4222 --name nats-server-1 --net $NETWORK_NAME -d "$NAT_SERVER_IMAGE" -p 4222
}

start_client()
{
  # start all clients
  port=8080
  for i in `seq 1 12`
  do
    docker run --rm -p $port:$port --name "nats-client-$i" --net $NETWORK_NAME -d "$NAT_CLIENT_IMAGE" -p $port
    port=$((port+1))
  done
  # TODO: ping clients to ensure it healthy
}

create_network
cleanup
start_cluster
#start
start_client
