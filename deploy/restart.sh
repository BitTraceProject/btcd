#!/bin/bash

CONTAINER_NAME=$1
DEPLOY_PWD=$PWD

function restart() {
  # cp 会直接覆盖旧的
  if [ "$CONTAINER_NAME" != "" ]; then
    cd $DEPLOY_PWD/.. || exit
    echo "rebuild and restart"
    docker cp ${OUTPUT_DIR}/btcd "${CONTAINER_NAME}":/bittrace/
    docker restart "$CONTAINER_NAME"
    cd $DEPLOY_PWD || exit
    exit 0
  else
    echo "[ERROR]CONTAINER_NAME not set"
  fi
}