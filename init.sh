#!/bin/bash

# network
# docker network create -d bridge help_others_network

# mysql
# docker pull mysql:8.0

# docker run -d --restart=always --name mysql --network help_others_network \
#   -e MYSQL_ROOT_PASSWORD=3515063609563648226 \
#   -v /usr/local/docker/help-others/data/mysql:/var/lib/mysql \
#   mysql:8.0

# redis


# server
docker build -t help-others-image -f ./assets/Dockerfile.server .

docker run -d --name help-others-container --network help_others_network \
    -p 80:80 --shm-size=256m --restart=always \
    -v /usr/local/docker/help-others/data/files:/go/server/files \
    help-others-image