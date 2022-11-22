#!/bin/bash

subnet_name=btcd_peer_network
temp_dir=temp
pwd=${PWD}

function clean() {
  source ${pwd}/peers/${CONTAINER_NAME}/.env
  docker stop ${CONTAINER_NAME}
  echo "stop ${CONTAINER_NAME}"
  docker rm ${CONTAINER_NAME}
  echo "rm ${CONTAINER_NAME}"
  sudo rm -rf ${pwd}/peers/${CONTAINER_NAME}
  echo "rm -rf ${pwd}/peers/${CONTAINER_NAME}"
}

clean
