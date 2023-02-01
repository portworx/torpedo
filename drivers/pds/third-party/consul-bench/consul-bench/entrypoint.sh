#!/bin/bash

consul-bench \
    -service ${SERVICE_NAME} \
    -register ${SERVICE_INSTANCES} \
    -flap-interval ${SERVICE_FLAP_SECONDS}s \
    -watchers ${SERVICE_WATCHERS}
