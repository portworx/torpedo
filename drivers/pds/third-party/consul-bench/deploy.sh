#!/bin/bash
make build
make push
export PDS_CLUSTER=$1
export PDS_NS=$2
VERSION=0.1.1 PDS_NS="${PDS_NS:=dev}" envsubst < <(cat bench.yaml) | kubectl apply -f -
