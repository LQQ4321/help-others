#!/bin/bash

chmod u+x ./restart.sh
docker stop help-others-container
docker rm help-others-container
docker rmi help-others-image
chmod u+x ./init.sh
./init.sh
docker ps