#!/bin/bash

apt-get update
apt-get install -y docker.io

docker pull docker.io/arpit1596/leader-election-client:latest
docker run -itd \
  --name leader-election-client \
  --restart always \
  -p 8080:8080 \
  docker.io/arpit1596/leader-election-client:latest
