#!/bin/bash
PDS_NS="${PDS_NS:=dev}" envsubst < <(cat bench.yaml) | kubectl delete -f -
