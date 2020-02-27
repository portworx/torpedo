#!/bin/bash -x

if [ -n "${VERBOSE}" ]; then
    VERBOSE="--v"
fi

if [ -z "${SCALE_FACTOR}" ]; then
    SCALE_FACTOR="10"
fi

if [ -z "${SCHEDULER}" ]; then
    SCHEDULER="k8s"
fi

if [ -z "${LOGLEVEL}" ]; then
    LOGLEVEL="debug"
fi

if [ -z "${CHAOS_LEVEL}" ]; then
    CHAOS_LEVEL="5"
fi
if [ -z "${MIN_RUN_TIME}" ]; then
    MIN_RUN_TIME="0"
fi

if [[ -z "$FAIL_FAST" || "$FAIL_FAST" = true ]]; then
    FAIL_FAST="--failFast"
else
    FAIL_FAST="-keepGoing"
fi

SKIP_ARG=""
if [ -n "$SKIP_TESTS" ]; then
    skipRegex=$(echo $SKIP_TESTS | sed -e 's/,/}|{/g')
    SKIP_ARG="--skip={$skipRegex}"
fi

FOCUS_ARG=""
if [ -n "$FOCUS_TESTS" ]; then
    focusRegex=$(echo $FOCUS_TESTS | sed -e 's/,/}|{/g')
    FOCUS_ARG="--focus={$focusRegex}"
fi

if [ -z "${UPGRADE_ENDPOINT_URL}" ]; then
    UPGRADE_ENDPOINT_URL=""
fi

if [ -z "${UPGRADE_ENDPOINT_VERSION}" ]; then
    UPGRADE_ENDPOINT_VERSION=""
fi

if [ -n "${PROVISIONER}" ]; then
    PROVISIONER="$PROVISIONER"
fi

if [ -z "${STORAGE_DRIVER}" ]; then
    STORAGE_DRIVER="pxd"
fi

CONFIGMAP=""
if [ -n "${CONFIG_MAP}" ]; then
    CONFIGMAP="${CONFIG_MAP}"
fi

if [ -z "${TORPEDO_IMG}" ]; then
    TORPEDO_IMG="portworx/torpedo:latest"
    echo "Using default torpedo image: ${TORPEDO_IMG}"
fi

if [ -z "${TIMEOUT}" ]; then
    TIMEOUT="720h0m0s"
    echo "Using default timeout of ${TIMEOUT}"
fi

if [ -z "$DRIVER_START_TIMEOUT" ]; then
    DRIVER_START_TIMEOUT="5m0s"
    echo "Using default timeout of ${DRIVER_START_TIMEOUT}"
fi

APP_DESTROY_TIMEOUT_ARG=""
if [ -n "${APP_DESTROY_TIMEOUT}" ]; then
    APP_DESTROY_TIMEOUT_ARG="--destroy-app-timeout=$APP_DESTROY_TIMEOUT"
fi

if [ -z "$STORAGENODE_RECOVERY_TIMEOUT" ]; then
    STORAGENODE_RECOVERY_TIMEOUT="35m0s"
    echo "Using default storage node recovery timeout of ${STORAGENODE_RECOVERY_TIMEOUT}"
fi

AZURE_TENANTID=""
if [ -n "$AZURE_TENANT_ID" ]; then
    AZURE_TENANTID="${AZURE_TENANT_ID}"
fi

AZURE_CLIENTID=""
if [ -n "$AZURE_CLIENT_ID" ]; then
    AZURE_CLIENTID="${AZURE_CLIENT_ID}"
fi

AZURE_CLIENTSECRET=""
if [ -n "$AZURE_CLIENT_SECRET" ]; then
    AZURE_CLIENTSECRET="${AZURE_CLIENT_SECRET}"
fi

for i in $@
do
case $i in
	--backup-driver)
	BACKUP_DRIVER=$2
	shift
	shift
	;;
esac
done

if [[ -z "$TEST_SUITE" || "$TEST_SUITE" == "" ]]; then
    TEST_SUITE='"bin/asg.test",
            "bin/autopilot.test",
            "bin/basic.test",
            "bin/reboot.test",
            "bin/upgrade.test",
            "bin/drive_failure.test",
            "bin/volume_ops.test",
            "bin/sched.test",
            "bin/node_decommission.test",'
else
  TEST_SUITE=$(echo \"$TEST_SUITE\" | sed "s/,/\",\n\"/g")","
fi
echo "Using list of test suite(s): ${TEST_SUITE}"


kubectl delete pod torpedo
state=`kubectl get pod torpedo | grep -v NAME | awk '{print $3}'`
timeout=0
while [ "$state" == "Terminating" -a $timeout -le 600 ]; do
  echo "Terminating torpedo..."
  sleep 1
  state=`kubectl get pod torpedo | grep -v NAME | awk '{print $3}'`
  timeout=$[$timeout+1]
done

if [ $timeout -gt 600 ]; then
  echo "Torpedo is taking too long to terminate. Operation timeout."
  describe_pod_then_exit
fi

TORPEDO_CUSTOM_PARAM_VOLUME=""
TORPEDO_CUSTOM_PARAM_MOUNT=""
CUSTOM_APP_CONFIG_PATH=""
if [ -n "${CUSTOM_APP_CONFIG}" ]; then
    kubectl create configmap custom-app-config --from-file=custom_app_config.yml=${CUSTOM_APP_CONFIG}
    CUSTOM_APP_CONFIG_PATH="/mnt/torpedo/custom_app_config.yml"
    TORPEDO_CUSTOM_PARAM_VOLUME="{ \"name\": \"custom-app-config-volume\", \"configMap\": { \"name\": \"custom-app-config\", \"items\": [{\"key\": \"custom_app_config.yml\", \"path\": \"custom_app_config.yml\"}] } }"
    TORPEDO_CUSTOM_PARAM_MOUNT="{ \"name\": \"custom-app-config-volume\", \"mountPath\": \"${CUSTOM_APP_CONFIG_PATH}\", \"subPath\": \"custom_app_config.yml\" }"
fi

TORPEDO_SSH_KEY_VOLUME=""
TORPEDO_SSH_KEY_MOUNT=""
if [ -n "${TORPEDO_SSH_KEY}" ]; then
    kubectl create secret generic key4torpedo --from-file=${TORPEDO_SSH_KEY}
    TORPEDO_SSH_KEY_VOLUME="{ \"name\": \"ssh-key-volume\", \"secret\": { \"secretName\": \"key4torpedo\", \"defaultMode\": 256 }}"
    TORPEDO_SSH_KEY_MOUNT="{ \"name\": \"ssh-key-volume\", \"mountPath\": \"/home/torpedo/\" }"
fi

TESTRESULTS_VOLUME="{ \"name\": \"testresults\", \"hostPath\": { \"path\": \"/mnt/testresults/\", \"type\": \"DirectoryOrCreate\" } }"
TESTRESULTS_MOUNT="{ \"name\": \"testresults\", \"mountPath\": \"/testresults/\" }"

VOLUMES="${TESTRESULTS_VOLUME}"

if [ -n "${TORPEDO_SSH_KEY_VOLUME}" ]; then
    VOLUMES="${VOLUMES},${TORPEDO_SSH_KEY_VOLUME}"
fi

VOLUME_MOUNTS="${TESTRESULTS_MOUNT}"

if [ -n "${TORPEDO_SSH_KEY_MOUNT}" ]; then
    VOLUME_MOUNTS="${VOLUME_MOUNTS},${TORPEDO_SSH_KEY_MOUNT}"
fi

if [ -n "${TORPEDO_CUSTOM_PARAM_VOLUME}" ]; then
    VOLUMES="${VOLUMES},${TORPEDO_CUSTOM_PARAM_VOLUME}"
fi

if [ -n "${TORPEDO_CUSTOM_PARAM_MOUNT}" ]; then
    VOLUME_MOUNTS="${VOLUME_MOUNTS},${TORPEDO_CUSTOM_PARAM_MOUNT}"
fi

K8S_VENDOR_KEY=""
K8S_VENDOR_VALUE=""
K8S_VENDOR_OPERATOR="Exists"
NODE_DRIVER="ssh"
if [ -n "${K8S_VENDOR}" ]; then
    case "$K8S_VENDOR" in
        kubernetes)
            K8S_VENDOR_KEY=node-role.kubernetes.io/master
            ;;
        rancher)
            K8S_VENDOR_KEY=node-role.kubernetes.io/controlplane
            K8S_VENDOR_OPERATOR="In"
            K8S_VENDOR_VALUE='values: ["true"]'
            ;;
        gke)
            # Run torpedo on worker node, where px installation is disabled. 
            K8S_VENDOR_KEY=px/enabled
            K8S_VENDOR_OPERATOR="In"
            K8S_VENDOR_VALUE='values: ["false"]'
            NODE_DRIVER="gke"
            ;;
        aks)
            # Run torpedo on worker node, where px installation is disabled. 
            K8S_VENDOR_KEY=px/enabled
            K8S_VENDOR_OPERATOR="In"
            K8S_VENDOR_VALUE='values: ["false"]'
            NODE_DRIVER="aks"
            ;;
        eks)
            # Run torpedo on worker node, where px installation is disabled.
            K8S_VENDOR_KEY=px/enabled
            K8S_VENDOR_OPERATOR="In"
            K8S_VENDOR_VALUE='values: ["false"]'
            ;;
    esac
else
    K8S_VENDOR_KEY=node-role.kubernetes.io/master
fi

echo "Deploying torpedo pod..."
cat <<EOF | kubectl create -f -
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: torpedo-account
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
   name: torpedo-role
rules:
  -
    apiGroups:
      # have access to everything except Secrets
      - "*"
    resources: ["*"]
    verbs: ["*"]
  - nonResourceURLs: ["*"]
    verbs: ["*"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: torpedo-role-binding
subjects:
- kind: ServiceAccount
  name: torpedo-account
  namespace: default
roleRef:
  kind: ClusterRole
  name: torpedo-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: Pod
metadata:
  name: torpedo
  labels:
    app: torpedo
spec:
  tolerations:
  - key: node-role.kubernetes.io/master
    operator: Equal
    effect: NoSchedule
  - key: node-role.kubernetes.io/controlplane
    operator: Equal
    value: "true"
  - key: node-role.kubernetes.io/etcd
    operator: Equal
    value: "true"
  - key: apps
    operator: Equal
    value: "false"
    effect: "NoSchedule"
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: ${K8S_VENDOR_KEY}
            operator: ${K8S_VENDOR_OPERATOR}
            ${K8S_VENDOR_VALUE}
  initContainers:
  - name: init-sysctl
    image: busybox
    imagePullPolicy: IfNotPresent
    securityContext:
      privileged: true
    command: ["sh", "-c", "mkdir -p /mnt/testresults && chmod 777 /mnt/testresults/"]
  containers:
  - name: torpedo
    image: ${TORPEDO_IMG}
    imagePullPolicy: Always
    command: [ "ginkgo" ]
    args: [ "--trace",
            "--timeout", "${TIMEOUT}",
            "$FAIL_FAST",
            "--slowSpecThreshold", "600",
            "$VERBOSE",
            "$FOCUS_ARG",
            "$SKIP_ARG",
            $TEST_SUITE
            "--",
            "--spec-dir", "../drivers/scheduler/k8s/specs",
            "--app-list", "$APP_LIST",
            "--scheduler", "$SCHEDULER",
            "--log-level", "$LOGLEVEL",
            "--node-driver", "$NODE_DRIVER",
            "--scale-factor", "$SCALE_FACTOR",
            "--minimun-runtime-mins", "$MIN_RUN_TIME",
            "--driver-start-timeout", "$DRIVER_START_TIMEOUT",
            "--chaos-level", "$CHAOS_LEVEL",
            "--storagenode-recovery-timeout", "$STORAGENODE_RECOVERY_TIMEOUT",
            "--provisioner", "$PROVISIONER",
            "--storage-driver", "$STORAGE_DRIVER",
            "--config-map", "$CONFIGMAP",
            "--custom-config", "$CUSTOM_APP_CONFIG_PATH",
            "--storage-upgrade-endpoint-url=$UPGRADE_ENDPOINT_URL",
            "--storage-upgrade-endpoint-version=$UPGRADE_ENDPOINT_VERSION",
            "$APP_DESTROY_TIMEOUT_ARG",
            "--backup-driver", "$BACKUP_DRIVER"
    ]
    tty: true
    volumeMounts: [${VOLUME_MOUNTS}]
    env:
    - name: TORPEDO_SSH_USER
      value: "${TORPEDO_SSH_USER}"
    - name: TORPEDO_SSH_PASSWORD
      value: "${TORPEDO_SSH_PASSWORD}"
    - name: TORPEDO_SSH_KEY
      value: "${TORPEDO_SSH_KEY}"
    - name: AZURE_TENANT_ID
      value: "${AZURE_TENANTID}"
    - name: AZURE_CLIENT_ID
      value: "${AZURE_CLIENTID}"
    - name: AZURE_CLIENT_SECRET
      value: "${AZURE_CLIENTSECRET}"
  volumes: [${VOLUMES}]
  restartPolicy: Never
  serviceAccountName: torpedo-account
EOF

echo "Waiting for torpedo to start running"

function describe_pod_then_exit {
  echo "Pod description:"
  kubectl describe pod torpedo
  exit 1
}

for i in $(seq 1 600) ; do
  printf .
  state=`kubectl get pod torpedo | grep -v NAME | awk '{print $3}'`
  if [ "$state" == "Error" ]; then
    echo "Error: Torpedo finished with $state state"
    describe_pod_then_exit
  elif [ "$state" == "Running" ]; then
    echo ""
    kubectl logs -f torpedo
  elif [ "$state" == "Completed" ]; then
    echo "Success: Torpedo finished with $state state"
    exit 0
  fi

  sleep 1
done

echo "Error: Failed to wait for torpedo to start running..."
describe_pod_then_exit
