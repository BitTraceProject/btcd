# This Dockerfile builds btcd from source and creates a small (55 MB) docker container based on alpine linux.
#
# Clone this repository and run the following command to build and tag a fresh btcd amd64 container:
#
# docker build . -t yourregistry/btcd
#
# You can use the following command to buid an arm64v8 container:
#
# docker build . -t yourregistry/btcd --build-arg ARCH=arm64v8
#
# For more information how to use this docker image visit:
# https://github.com/btcsuite/btcd/tree/master/docs
#
# 8333  Mainnet Bitcoin peer-to-peer port
# 8334  Mainet RPC port

ARG ARCH=amd64
FROM docker.io/library/golang:1.18.3

ARG ARCH
ENV GO111MODULE=on
ENV GOARCH=$GOARCH
ENV GOOS=linux

ADD . /bitlog
WORKDIR /bitlog

RUN set -ex \
  && go env -w GO111MODULE=on \
  && go env -w GOPROXY=https://goproxy.cn,direct \
  && go mod tidy \
  && go install -v . ./cmd/...

VOLUME ["/root/.btcd"]


# for testnet
EXPOSE 18555 18556
# for mainnet
EXPOSE 8333 8334
# for simnet
EXPOSE 18333 18334
# for signet
EXPOSE 38333 38334

ENTRYPOINT ["btcd"]
