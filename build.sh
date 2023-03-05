#!/bin/bash

export GOOS="linux" # windows darwin linux
export OUTPUT_DIR="./output"

function build() {
  mkdir -p $OUTPUT_DIR
  set -o errexit
  export GO111MODULE=on
  go env -w GO111MODULE=on
  go env -w GOPROXY=https://goproxy.cn,direct
  go mod tidy
  go build -v -o "$OUTPUT_DIR/btcd" .
  echo "build successfully!"
}

build
