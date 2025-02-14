# Export required environment variables
export WORKSPACE="/opt/workspace"
export KUBECONFIG="$WORKSPACE/kubeconfig"
export TORPEDO_TAR="torpedo_master.tar"
export TORPEDO_IMG="localhost/torpedo:master"

# Change below variables as needed
## Tests to run
export FOCUS_TESTS="SingleVMLiveMigration,MultipleParallelLiveMigration,LiveMigrationsOfVMsInALoop,ColdAddNewDiskToKubevirtVM,LMAfterColdAddDiskToVM,LMBeforeColdAddDiskToVM,PxKillAfterColdAddDiskToVM,KillPxOnSourceNodeDuringMigration,KillPxOnDestNodeDuringMigration,DeleteMigrationObjectDuringMigration,RepeatedDeleteMigrationObjectDuringMigration"
export APP_LIST="kubevirt-debian-fio-minimal"
export TEST_DESC="px-ocp-kubevirt-all"
export PROVISIONER="csi"
export STORAGE_DRIVER="pxd"
export SCHEDULER="openshift"
export SCALE_FACTOR="1"
export KUBEVIRT_VOL_TYPE=""
#select either fada-raw, pxe-raw, or leave it blank for sharedV4 volumes

#Modifiable variables
export CUSTOM_SPEC_DIR="$WORKSPACE/custom_spec"
export TIMEOUT="720h0m0s"
export PURE_SAN_TYPE="ISCSI"
export LICENSE_EXPIRY_TIMEOUT_HOURS="1h0m0s"
export METERING_INTERVAL_MINS="10m0s"
export SLOW_SPEC_THRESHOLD="600"
export DEPLOY_PDS_APPS="false"
export MAX_STORAGE_NODES_PER_AZ="1"
export FAIL_ON_PX_POD_RESTARTCOUNT="false"
export DRIVER_START_TIMEOUT="30m0s"
export CHAOS_LEVEL="5"
export STORAGENODE_RECOVERY_TIMEOUT="35m0s"
export LOG_LEVEL="debug"

#DO NOT CHANGE variables
export LAUNCHER_SCRIPT="deploy-ssh-ocp-standalone.sh"
export KUBEVIRT_VM_PWD="Password1"
export HYPER_CONVERGED="true"
export IS_OCP="true"
export SECRET_TYPE="k8s"
export SPEC_DIR="../drivers/scheduler/k8s/specs"
export TORPEDO_SKIP_SYSTEM_CHECKS="true"
export SECURITY_CONTEXT="false"
export K8S_VENDOR="kubernetes"

# Load the Torpedo image
podman load -i "$WORKSPACE/$TORPEDO_TAR"

# List images to verify
podman images

#enable podman socket
sudo systemctl enable podman.socket

# Change directory to workspace
cd "$WORKSPACE"

# Start local server to host debian image endpoint for kubevirt VM pvc
export HOST_IP=$(hostname -i | awk '{print $1}')
nohup python3 -m http.server --bind "$HOST_IP" 8000 --directory "$WORKSPACE" > local_server.out &
SERVER_PID=$!

# Trap to ensure the server is stopped when the script exits
trap "kill $SERVER_PID" EXIT

echo "HTTP server running on $HOST_IP with PID $SERVER_PID"

# Copy Torpedo data into the workspace
podman run --rm --privileged -v "$WORKSPACE:$WORKSPACE" --entrypoint '' "$TORPEDO_IMG" cp -r /torpedo "$WORKSPACE"

#Keep below line commented out. Only uncomment for debugging purposes
cp /opt/workspace/deploy-ssh-ocp-standalone.sh "$WORKSPACE/torpedo/deployments/"

# Run the Torpedo test suite
podman run --rm -t --privileged --net=host \
-v "$WORKSPACE/torpedo/deployments:/deployments" \
-v "$KUBECONFIG":"$KUBECONFIG" \
-v /var/run/podman/podman.sock:/var/run/podman/podman.sock \
-e KUBECONFIG="$KUBECONFIG" \
-e TORPEDO_IMG="$TORPEDO_IMG" \
-e VERBOSE=true \
-e TEST_SUITE= \
-e FOCUS_TESTS="$FOCUS_TESTS" \
-e SCALE_FACTOR=1 \
-e SCHEDULER="$SCHEDULER" \
-e K8S_VENDOR="$K8S_VENDOR" \
-e APP_LIST="$APP_LIST" \
-e FAIL_FAST=true \
-e PROVISIONER="$PROVISIONER" \
-e HOST_IP="$HOST_IP" \
-e KUBEVIRT_VM_PWD="$KUBEVIRT_VM_PWD" \
-e KUBEVIRT_VOL_TYPE="$KUBEVIRT_VOL_TYPE" \
-e TEST_DESC="$TEST_DESC" \
-e CUSTOM_SPEC_DIR="$CUSTOM_SPEC_DIR" \
--entrypoint /bin/sh \
lachlanevenson/k8s-kubectl \
/deployments/$LAUNCHER_SCRIPT
