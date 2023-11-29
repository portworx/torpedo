#!/bin/bash

# Specify the Kubernetes namespace
NAMESPACE="pds-system"

# Create a directory to store individual pod log files
LOGS_DIR="pod_logs"
mkdir -p ${LOGS_DIR}

# Get the list of pod names in the specified namespace
POD_NAMES=$(kubectl get pods --namespace=${NAMESPACE} -o jsonpath='{.items[*].metadata.name}')

# Loop through each pod and retrieve its logs
for POD_NAME in ${POD_NAMES}; do
    POD_LOG_FILE="${LOGS_DIR}/${POD_NAME}_logs.txt"
   
    # Check if the pod name contains "prometheus"
    if [[ ${POD_NAME} == *"prometheus"* ]]; then
        echo "Getting logs for Prometheus server container"
        echo "Logs for Pod: ${POD_NAME}" >> ${POD_LOG_FILE}
        kubectl logs ${POD_NAME} -c "pds-tc-prometheus-server"  --namespace=${NAMESPACE} >> ${POD_LOG_FILE}
        echo "-------------------------------------" >> ${POD_LOG_FILE}
        
        echo "Logs for Pod ${POD_NAME} saved to ${POD_LOG_FILE}"
        continue
    fi
 
    echo "Logs for Pod: ${POD_NAME}" >> ${POD_LOG_FILE}
    kubectl logs ${POD_NAME} --namespace=${NAMESPACE} >> ${POD_LOG_FILE}
    echo "-------------------------------------" >> ${POD_LOG_FILE}
    
    echo "Logs for Pod ${POD_NAME} saved to ${POD_LOG_FILE}"
done

# Zip the directory containing log files
ZIP_FILE="pod_logs.zip"
zip -r ${ZIP_FILE} ${LOGS_DIR}

# Move the zip file to a desired location if needed
# mv ${ZIP_FILE} /path/to/destination/

echo "Log files zipped and saved to ${ZIP_FILE}"

