#!/bin/sh

apt-get update
apt-get install -y docker.io

mkdir /raft-cluster
docker pull docker.io/arpit1596/leader-election:latest
docker run -itd \
  --name leader-election \
  --restart always \
  -e SERVER_ID="10.0.1.10:50081" \
  -e BOOTSTRAP_CLUSTER="true" \
  -e SERVER_ADDRESS="10.0.1.10" \
  -e PORT="50081" \
  -v /raft-cluster:/raft-cluster \
  -p 10.0.1.10:50081:50081 \
   -p 10.0.1.10:50082:50082 \
  docker.io/arpit1596/leader-election:latest
