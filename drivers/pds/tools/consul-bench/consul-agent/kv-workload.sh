#!/bin/bash

export CONSUL_HTTP_TOKEN=${AGENT_TOKEN}

while :
do
    consul kv get -detailed pds/loadtest/foo
    consul kv put pds/loadtest/foo "$(date)"
    consul kv get -detailed pds/loadtest/foo

    sleep 1

    curl --header "X-Consul-Token: ${CONSUL_HTTP_TOKEN}" \
        localhost:8500/v1/kv/pds/loadtest/foo
    curl --header "X-Consul-Token: ${CONSUL_HTTP_TOKEN}"  \
        --request PUT \
        --data "$(date)" \
        localhost:8500/v1/kv/pds/loadtest/foo
    sleep 1
done
