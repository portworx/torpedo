#!/bin/bash

# Specify the Kubernetes namespace
NAMESPACE="pds-system"

# Create a directory to store individual pod log files
LOGS_DIR="pds_container_logs"
mkdir -p ${LOGS_DIR}

# Get the list of pod names in the specified namespace
POD_NAMES=$(kubectl get pods --namespace=${NAMESPACE} -o jsonpath='{.items[*].metadata.name}')

# Loop through each pod and retrieve logs for all containers
for POD_NAME in ${POD_NAMES}; do
    POD_LOG_DIR="${LOGS_DIR}/${POD_NAME}"
    mkdir -p ${POD_LOG_DIR}

    # Get the list of containers in the pod
    CONTAINER_NAMES=$(kubectl get pod ${POD_NAME} --namespace=${NAMESPACE} -o jsonpath='{.spec.containers[*].name}')

    # Loop through each container and retrieve its logs
    for CONTAINER_NAME in ${CONTAINER_NAMES}; do
        CONTAINER_LOG_FILE="${POD_LOG_DIR}/${CONTAINER_NAME}_logs.txt"

        echo "Logs for Pod: ${POD_NAME}, Container: ${CONTAINER_NAME}" >> ${CONTAINER_LOG_FILE}
        kubectl logs ${POD_NAME} -c ${CONTAINER_NAME} --namespace=${NAMESPACE} >> ${CONTAINER_LOG_FILE}
        echo "-------------------------------------" >> ${CONTAINER_LOG_FILE}

        echo "Logs for Pod ${POD_NAME}, Container ${CONTAINER_NAME} saved to ${CONTAINER_LOG_FILE}"
    done
done

# Zip the directory containing log files
ZIP_FILE="pds_container_logs.zip"
zip -r ${ZIP_FILE} ${LOGS_DIR}

# Move the zip file to a desired location if needed
# mv ${ZIP_FILE} /path/to/destination/

echo "Log files zipped and saved to ${ZIP_FILE}"

