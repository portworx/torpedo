#!/bin/bash

envsubst < "/tmp/agent.json" > "/config/agent.json"
consul agent -node=$(hostname) -retry-join=${PDS_CLUSTER}-${PDS_NS}-0 -data-dir=/var/consul -server=false -config-dir=/config
