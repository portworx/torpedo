# HOW TO DEPLOY TORPEDO AND RUN TESTS IN YOUR OCP LAB

This document provides instructions on how to deploy torpedo in your OpenShift (OCP) in standalone mode 
(without the use of Jenkins or spawn).  

## Pre-requisites:
* An OpenShift (OCP) deployment with a minimum of 3 nodes in the Kubernetes cluster
* Portworx installed on the Kubernetes cluster
* ssh access to the master node to copy tar image 
* `kubeconfig` file for the cluster

## Artifacts Needed - Download the artifacts from a shared location to be provided by bnagar@purestorage.com or dbhatnagar@purestorage.com
- Torpedo.tar
- fio-data-validation.qcow2 (For FADA-Raw)
- launch-torpedo.sh
- custom_spec (directory) 
  - This directory contains supporting files to deploy the PVC, Storage Class and VMs for Torpedo tests. 
  - The use of this directory is optional while running torpedo tests
  - You can edit these files if you wish to use your own specs, but keep the files with same names and directory structure.

### Delete the text between start and end INTERNAL_ONLY tags below before sharing this README with external sources
#### If you are Pure employee
#### >>> start INTERNAL_ONLY <<<

1. Download the app image and custom_spec directory from 
   - https://drive.google.com/drive/folders/1tbSLcIqt45P11Bcz1-J8Aa0KTErKkuwA?usp=drive_link
   - Chose either SharedV4 Image name or FADA Raw image based on your setup 
   - Download the custom_spec directory to give to a customer
2. Build `torpedo` tar -
- Build torpedo tar from master branch - https://devops-jenkins.pwx.purestorage.com/job/Builds/job/torpedo/job/torpedo-branch/
- Use the image from the artifactory in the commands below. 

```shell
IMAGE="pure-artifactory.dev.purestorage.com/px-docker-prod-virtual/portworx/torpedo:master"
docker pull $IMAGE 
docker tag $IMAGE localhost/torpedo:master && docker save -o torpedo_master.tar $IMAGE
```
3. Get `launch-torpedo.sh` from https://github.com/pure-px/torpedo/tree/master/deployments/launch-torpedo.sh

#### >>> end INTERNAL_ONLY <<<

## Copy the artifacts to the Master node and Run Torpedo tests

1. From a machine which has access to a master node, copy the provided artifacts to a directory on Master Node.
   Below example uses key based authentication to copy files.
   It also creates a workspace directory on Master Node. 

   ```shell
   export WORKSPACE="/opt/workspace/"
   export MASTER_NODE_IP="10.38.15.132"
   export APP_IMAGE="fio-data-validation.qcow2"
   ssh -i ~/id_rsa core@$MASTER_NODE_IP "sudo mkdir -p $WORKSPACE; sudo mkdir -p $WORKSPACE/images; sudo chmod -R 777 $WORKSPACE"
   scp -i ~/id_rsa torpedo_master.tar core@$MASTER_NODE_IP:$WORKSPACE/torpedo_master.tar
   scp -i ~/id_rsa launch-torpedo.sh core@$MASTER_NODE_IP:$WORKSPACE/launch-torpedo.sh
   scp -i ~/id_rsa -r custom_spec core@$MASTER_NODE_IP:$WORKSPACE/custom_spec
   scp -i ~/id_rsa $APP_IMAGE core@$MASTER_NODE_IP:$WORKSPACE/images/$APP_IMAGE
   scp -i ~/id_rsa kubeconfig core@$MASTER_NODE_IP:$WORKSPACE/kubeconfig
   ```

2. SSH to Master Node and switch to root user. 
   ```shell
   ssh -i ~/id_rsa core@$MASTER_NODE_IP
   sudo su
   export WORKSPACE="/opt/workspace/"
   cd $WORKSPACE
   ```
3. Edit the `launch-torpedo.sh` file and configure the values according to your environment
Update the exported variables according to your environment. 

   ```shell
   vi launch-torpedo.sh
   chmod 755 launch-torpedo.sh 
   ```
   <TODO: To be updated before sending out>
   The `FOCUS_TESTS` variable specifies which tests will be run. Below is a list of available tests you can run: 
   - AddNewDiskToKubevirtVM
   - KubeVirtLiveMigration
   - PxKillBeforeAddDiskToVM
   - PxKillAfterAddDiskToVM
   - KubevirtVMVolHaIncrease
   - KubevirtVMVolHaDecrease
   - LiveMigrationBeforeAddDisk
   - AddDiskAndLiveMigrate
   - KubeVirtPvcAndPoolExpandWithAutopilot
   - UpgradeOCPAndValidateKubeVirtApps
   - RebootRootDiskAttachedNode
   - ParallelAddDiskToVM
   - MultipleKubeVirtLiveMigration
   - AddDiskAndLiveMigrateMultipleVm
   - LiveMigrationBeforeAddDiskMultipleVm
   - MultipleVMVolHaIncrease
   - MultipleVMVolHaDecrease
   - LiveMigrateWhileNodeInMaintenance
   - LiveMigrateCordonNonReplicaNode
   - StopPxOnNodeWhereVMIsProvisioned
   - RestartPXAndCheckIfVmBindMount
   - FillVMRootDisk
   - SingleVMLiveMigration
   - SingleVMLiveMigrationStorkUpgrade
   - MultipleParallelLiveMigration
   - LiveMigrationsOfVMsInALoop
   - AddNewRawDiskToKubevirtVM
   - LMAfterRawDiskAddToVM
   - LMBeforeRawDiskAddToVM
   - LMAndAddRawDiskToVMInALoop
   - PxKillAfterAddRawDiskToVM
   - AddDiskKillPxLMAgainAddDisk
   - KillPxOnSourceNodeDuringMigration
   - KillPxOnDestNodeDuringMigration
   - KillPxOnSrcAndDestNodesDuringLM
   - DeleteMigrationObjectDuringMigration
   - RepeatedDeleteMigrationObjectDuringMigration
   - AddNewMixedDiskToKubevirtVMAndLM
   - ResizePvcAndLiveMigrateVMs
   - AddNewHotPlugDiskToKubevirtVM
   - RebootNodeAfterAddNewHotPlugDiskToKubevirtVM
   - PxRestartAfterAddNewHotPlugDiskToKubevirtVM
   - VMLiveMigrationHavingMultipleSC
   - LMAfterAddNewHotPlugDiskToKubevirtVM
   - Add26NewHotPlugDiskToKubevirtVM
   - PxRestartDuringAddNewHotPlugDiskToKubevirtVM
   - RebootNodeDuringAddNewHotPlugDiskToKubevirtVM
   - RebootSourceNodeDuringMigration
   - LMAfterFillingDisksOfKubevirtVM
   - AddAndRemoveNewHotPlugDiskToKubevirtVM
   
   You can specify multiple tests in a comma separated format in the `launch-torpedo.sh` file by updating the `FOCUS_TESTS` variable
   ```shell
   export FOCUS_TESTS="SingleVMLiveMigration,MultipleParallelLiveMigration,...,<YOUR_TEST_NAME>"
   ```
4. [OPTIONAL] Using the `custom_spec` directory:
   - To use the `custom_spec` directory when running the test, edit the `CUSTOM_SPEC_DIR` variable in the `launch-torpedo.sh` script
   - point it to the `custom_spec` directory in the `WORKSPACE` 
   - This directory is mounted inside the Torpedo container during its creation
5. Run the `launch-torpedo.sh` script to deploy Torpedo and run the specified tests
   ```shell
   ./launch-torpedo.sh
   ```
   - This script will create a local HTTP server on the master node running on `http://$HOST_IP:8000` which will host the images in `WORKSPACE`. This is needed to host image for creating the PVC using `cdi.kubevirt.io/storage.import.endpoint`**