#!/bin/bash

CONTAINER_NAME=$1
DEPLOY_PWD=$PWD
OUTPUT_DIR=$PWD/../output

function restart() {
  set -x
  # cp 会直接覆盖旧的
  if [ "$CONTAINER_NAME" != "" ]; then
    echo "rebuild and restart"
    cd $DEPLOY_PWD/.. || exit
    bash $DEPLOY_PWD/../build.sh
    docker cp ${OUTPUT_DIR}/btcd "${CONTAINER_NAME}":/bittrace/
    docker restart "$CONTAINER_NAME"
    cd $DEPLOY_PWD || exit
    exit 0
  else
    echo "[ERROR]CONTAINER_NAME not set"
  fi
}

restart
