#!/bin/bash -x

DAEMONSET_EXISTS="false"
if kubectl get daemonset debug -n kube-system > /dev/null 2>&1; then
    DAEMONSET_EXISTS="true"
    kubectl delete daemonset debug -n kube-system
fi
echo "daemonset exists: [$DAEMONSET_EXISTS]"

kubectl apply -f "$DEBUG_YAML_PATH"

TORPEDO_NODE=$(kubectl get pod torpedo -o jsonpath='{.spec.nodeName}')

timeout=300
interval=10
while [ $timeout -gt 0 ]; do
    if kubectl get daemonset debug -n kube-system -o jsonpath='{.status.numberReady}' | grep -q '^[1-9][0-9]*$'; then
        break
    fi
    echo "waiting for debug DaemonSet to run"
    sleep $interval
    (( timeout -= interval ))
done

if [ "$timeout" -le 0 ]; then
    echo "exited due to timeout while waiting for debug DaemonSet to reach Running state"
    exit 1
fi

mapfile -t NODES < <(kubectl get pods -l name=debug -o wide -n kube-system | awk '{print $7}' | tail -n +2)

TORPEDO_NODE_FOUND=false

for node in "${NODES[@]}"; do
    if [ "$node" == "$TORPEDO_NODE" ]; then
        TORPEDO_NODE_FOUND=true
        break
    fi
done

echo "found torpedo node: [$TORPEDO_NODE_FOUND]"

if [ "$TORPEDO_NODE_FOUND" == "true" ]; then
    DEBUG_POD_ON_TORPEDO_NODE=$(kubectl get pods -l name=debug -o wide -n kube-system | grep "$TORPEDO_NODE" | awk '{print $1}' | head -n 1)
    kubectl cp kube-system/"$DEBUG_POD_ON_TORPEDO_NODE":testresults/junit_basic.xml "$WORKSPACE_MOUNT_PATH/testresults/junit_basic.xml"
    echo "copied junit_basic.xml from [$DEBUG_POD_ON_TORPEDO_NODE] on torpedo node [$TORPEDO_NODE] to [$WORKSPACE_MOUNT_PATH/testresults/junit_basic.xml]"
fi

#mapfile -t PORTWORX_PODS < <(kubectl get po -l name=portworx -n kube-system -o custom-columns=:metadata.name)
#echo "PORTWORX_PODS: [${PORTWORX_PODS[*]}]"
#
#for pod in "${PORTWORX_PODS[@]}"; do
#    kubectl exec "$pod" -n kube-system -- pxctl sv diags -a -c
#    echo "executed 'pxctl sv diags -a -c' on pod $pod"
#done
#
#AVAILABLE_MOUNT_PATHS=$(kubectl get pod "$DEBUG_POD" -n kube-system -o jsonpath='{.spec.containers[*].volumeMounts[*].mountPath}')
#echo "AVAILABLE_MOUNT_PATHS: [$AVAILABLE_MOUNT_PATHS]"
#
#mapfile -t MOUNT_PATHS_LIST < <(tr ':' '\n' <<< "$MOUNT_PATHS")
#
#for node in "${NODES[@]}"; do
#    DEBUG_POD=$(kubectl get pods -l name=debug -o wide -n kube-system | grep "$node" | awk '{print $1}' | head -n 1)
#    POD_STATUS=$(kubectl get pod "$DEBUG_POD" -n kube-system -o jsonpath='{.status.phase}')
#    if [ "$POD_STATUS" != "Running" ]; then
#        echo "cannot to copy files as pod [$DEBUG_POD] on node [$node] is not running"
#        continue
#    fi
#    MOUNT_PATH_ADDED=false
#    TAR_FILENAME="${node}.tar.gz"
#    TAR_COMMAND="tar cfvz /tmp/$TAR_FILENAME --absolute-names --warning=no-file-changed"
#    for mountPath in "${MOUNT_PATHS_LIST[@]}"; do
#        if kubectl exec -n kube-system "$DEBUG_POD" -- test -d "$mountPath" || test -f "$mountPath"; then
#            TAR_COMMAND+=" $mountPath"
#            MOUNT_PATH_ADDED=true
#        fi
#    done
#    if [ "$MOUNT_PATH_ADDED" = true ]; then
#        TIMESTAMP=$(date "+%Y%m%d%H%M%S")
#        kubectl exec -n kube-system "$DEBUG_POD" -- chroot /proc/1/root /bin/bash -c "journalctl -u kubelet > /var/cores/kubelet-$TIMESTAMP.log"
#        kubectl exec -n kube-system "$DEBUG_POD" -- bash -c "$TAR_COMMAND"
#        ARCHIVE_PATH="$WORKSPACE_MOUNT_PATH/archives/cluster/$node"
#        mkdir -p "$ARCHIVE_PATH"
#        kubectl cp kube-system/"$DEBUG_POD:tmp/$TAR_FILENAME" "$ARCHIVE_PATH/${TAR_FILENAME}"
#        echo "copied tar file [$TAR_FILENAME] from pod [$DEBUG_POD] on node [$node] to [$ARCHIVE_PATH/$TAR_FILENAME]"
#    else
#        echo "no valid mount paths found for pod [$DEBUG_POD] on node [$node]"
#    fi
#done

if [ "$DAEMONSET_EXISTS" = "false" ]; then
    kubectl delete -f "$DEBUG_YAML_PATH"
fi
