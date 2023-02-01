#!/bin/bash

CONTAINER_NAME=$1

function restart() {
  if [ "$RESTART_FLAG" != "" ]; then
    echo "[RESTART]${CONTAINER_NAME}"
    docker restart ${CONTAINER_NAME}
    exit 0
  else
    echo "[ERROR]CONTAINER_NAME not set"
  fi
}