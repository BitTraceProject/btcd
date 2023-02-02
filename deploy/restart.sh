#!/bin/bash

CONTAINER_NAME=$1
DEPLOY_PWD=$PWD

function restart() {
  # cp 会直接覆盖旧的
  cd $DEPLOY_PWD/.. || exit
  if [ "$CONTAINER_NAME" != "" ]; then
    echo "rebuild and restart"
    docker cp ${OUTPUT_DIR}/btcd "${CONTAINER_NAME}":/bittrace/
    docker restart "$CONTAINER_NAME"
    exit 0
  else
    echo "[ERROR]CONTAINER_NAME not set"
  fi
}