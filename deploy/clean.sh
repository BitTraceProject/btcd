#!/bin/bash

CONTAINER_NAME=$1
pwd=${HOME}/.bittrace
peer_pwd=${pwd}/peers

function clean() {
  source ${peer_pwd}/${CONTAINER_NAME}/.env
  docker stop ${CONTAINER_NAME}
  echo "stop ${CONTAINER_NAME}"
  docker rm ${CONTAINER_NAME}
  echo "rm ${CONTAINER_NAME}"
  sudo rm -rf ${peer_pwd/${CONTAINER_NAME}
  echo "rm -rf ${peer_pwd}/${CONTAINER_NAME}"
}

clean $1