#!/bin/bash

for ns in `kubectl get ns -l creator=torpedo | grep Active | awk '{print $1}'`; do
    echo "Cleaning up torpedo from namespace: ${ns}"
    kubectl delete namespace "${ns}"
done
