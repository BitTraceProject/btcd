#!/bin/bash

subnet_name=btcd_peer_network
temp_dir=temp
pwd=${PWD}

function clean() {
  docker stop ${CONTAINER_NAME}
  docker rm ${CONTAINER_NAME}
  sudo rm -rf ${pwd}/peers/${CONTAINER_NAME}
}

clean
