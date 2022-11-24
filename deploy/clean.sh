#!/bin/bash

pwd=${HOME}/.bittrace
peer_pwd=${pwd}/peers

CONTAINER_NAME=$1
source ${peer_pwd}/${CONTAINER_NAME}/.env
docker stop ${CONTAINER_NAME}
echo "stop ${CONTAINER_NAME}"
docker rm ${CONTAINER_NAME}
echo "rm ${CONTAINER_NAME}"
sudo rm -rf ${peer_pwd}/${CONTAINER_NAME}
echo "rm -rf ${peer_pwd}/${CONTAINER_NAME}"