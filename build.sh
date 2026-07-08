#!/bin/bash

prog=$(realpath "$0")
root=$(dirname "$prog")

rc=0

function log_info { echo "[INFO] $*"; }
function log_error {
  echo "[ERRO] $*"
  rc=1
}

function do_echo {
  echo "[CMD ] $*"
  TIMEFORMAT="[TIME] took %3lR"
  time "$@"
  code=$?
  if [ $code -ne 0 ]; then
    rc=$code
    log_error "return code $rc"
  fi
}

function usage {
  name=$(basename "$prog")
  log_info "Usage: ${name} [-h] [-b] [-t] [-l] [-g]"
  log_info "  -h: Show this help message"
  log_info "  -b: Build the modwhy binary"
  log_info "  -t: Run go tests"
  log_info "  -l: Run golangci-lint"
  log_info "  -g: Format with gofumpt"
}

cd "$root" || exit 1

if [ "$#" -eq "0" ]; then
  usage
fi

while [ "$#" -gt "0" ]; do
  arg=$1
  shift

  case ${arg} in
  -h)
    usage
    ;;
  -b)
    do_echo go build -o bin/modwhy ./cmd/modwhy
    log_info "Run with ./bin/modwhy"
    ;;
  -t)
    do_echo go test -cover ./...
    ;;
  -l)
    do_echo go tool golangci-lint run
    ;;
  -g)
    do_echo go fix ./...
    do_echo go tool gofumpt -w .
    ;;
  *)
    log_error "Unknown argument ${arg}"
    usage
    ;;
  esac
done

exit "${rc}"
