#!/bin/bash

export GOOS="linux" # windows darwin linux
export OUTPUT_DIR="./output"
export PACKAGE_NAME="btcd"

CONTAINER_NAME=$1 # 不为空代表 rebuild，其余情况都是 build

function build() {
    mkdir -p $OUTPUT_DIR
    set -o errexit
    export GO111MODULE=on
    go env -w GO111MODULE=on
    go env -w GOPROXY=https://goproxy.cn,direct
    go mod tidy
    go build -v -o "$OUTPUT_DIR/btcd" .
    go build -v -o "$OUTPUT_DIR/" ./cmd/...
    echo "build successfully!"
}

function buildContainer() {
    build

    docker build . -t bittrace/peer_btcd
    echo "build container successfully!"
}

function rebuildContainer() {
    build

    # cp 会直接覆盖旧的
    docker cp ${OUTPUT_DIR}/* "${CONTAINER_NAME}":/bittrace/
    echo "rebuild container successfully!"
}

if [ "$CONTAINER_NAME" != "" ]; then
    echo "[REBUILD]${CONTAINER_NAME}"
    rebuildContainer
else
    # default
    echo "[BUILD]"
    buildContainer
fi