#!/bin/bash

subnet_name=peer_btcd_network
pwd=${HOME}/.bittrace
peer_dir=${pwd}/peers
tmpl_dir=${pwd}/tmpl
temp_dir=${pwd}/.temp

function precheck() {
  if [ ! -f "${tmpl_dir}/.env.tmpl" ]; then
    exitWithError ".env tmpl not prepare"
  fi
  if [ ! -f "${tmpl_dir}/btcd.conf.tmpl" ]; then
    exitWithError "btcd.conf.tmpl not prepared"
  fi
  if [ ! -f "${tmpl_dir}/docker-compose.yaml.tmpl" ]; then
    exitWithError "docker-compose.yaml not prepared"
  fi
  source "${tmpl_dir}/.env.tmpl"
  if [ -z "${CONTAINER_NAME}" ]; then
    exitWithError "peer name not set, please set"
  fi
  if [ -d "${peer_dir}/${CONTAINER_NAME}" ]; then
    exitWithError "peer container dir has existed, please rm or rename peer"
  fi
}

function prepare() {
  cd "${pwd}"/ || exit 0
  rm -rf "${temp_dir}"
  mkdir "${temp_dir}"
}

function bootstrap() {
  infoln "up peer container"
  docker-compose up -d
}

function clean() {
    pwd=${HOME}/.bittrace
    peer_dir=${pwd}/peers
    CONTAINER_NAME=$1
    source ${peer_dir}/${CONTAINER_NAME}/.env
    docker stop ${CONTAINER_NAME}
    echo "stop ${CONTAINER_NAME}"
    docker rm ${CONTAINER_NAME}
    echo "rm ${CONTAINER_NAME}"
    sudo rm -rf ${peer_dir}/${CONTAINER_NAME}
    echo "rm -rf ${peer_dir}/${CONTAINER_NAME}"
}

function exitWithError() {
  errorMsg=$1
  if [ -n "${errorMsg}" ]; then
    errorln "${errorMsg}"
  fi
  if [ -d "${temp_dir}" ]; then
    infoln "clean temp files"
    sudo rm -rf "${pwd}"/"${temp_dir}"/
  fi
  clean ${CONTAINER_NAME}
  exit 0
}

function main() {
  set -x
  infoln "precheck process"
  precheck

  infoln "prepare process"
  prepare

  infoln "copy tmpl files"
  cp "${tmpl_dir}/.env.tmpl" "${temp_dir}/.env"
  infoln "copy btcd.conf tmpl"
  cp "${tmpl_dir}/btcd.conf.tmpl" "${temp_dir}/btcd.conf"
  infoln "copy docker-compose.yaml tmpl"
  cp "${tmpl_dir}/docker-compose.yaml.tmpl" "${temp_dir}/docker-compose.yaml"

  infoln "copy tmpl to peer ${CONTAINER_NAME}"
  mkdir -p "${peer_dir}/${CONTAINER_NAME}/.btcd/"
  cp "${temp_dir}/.env" "${peer_dir}/${CONTAINER_NAME}/"
  cp "${temp_dir}/btcd.conf" "${peer_dir}/${CONTAINER_NAME}/.btcd/"
  cp "${temp_dir}/docker-compose.yaml" "${peer_dir}/${CONTAINER_NAME}/"

  infoln "source env"
  cd "${peer_dir}/${CONTAINER_NAME}" || exit
  source .env

  infoln "bootstrap peer"
  bootstrap

  rm -rf "${temp_dir}"
  infoln "bootstrap success"
  exitWithError
}

function infoln() {
    echo "===[info]: ${1}"
}

function errorln() {
    echo "===[error]: ${1}"
}

main
