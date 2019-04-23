#!/bin/bash -x

if [ -n "${VERBOSE}" ]; then
    VERBOSE="--v"
fi

if [ -z "${SCALE_FACTOR}" ]; then
    SCALE_FACTOR="3"
fi

if [ -n "${SKIP_TESTS}" ]; then
    SKIP_ARG="--skip=$SKIP_TESTS"
fi

if [ -n "${FOCUS_TESTS}" ]; then
    focusRegex=$(echo $FOCUS_TESTS | sed -e 's/,/}|{/g')
    FOCUS_ARG="--focus={$focusRegex}"
fi

if [ -z "${TORPEDO_IMAGE}" ]; then
    TORPEDO_IMAGE="portworx/torpedo:late"
    printf "Using default torpedo image: ${TORPEDO_IMAGE}\n"
fi

if [ -z "${NOMAD_API_ADDR}" ]; then
    NOMAD_API_ADDR="127.0.0.1"
    printf "Using default HTTP API address\n"
else
    NOMAD_API_ADDR="$NOMAD_API_ADDR"
fi

if [ -n "${TORPEDO_SSH_USER}" ]; then
    TORPEDO_SSH_USER="$TORPEDO_SSH_USER"
fi

if [ -n "${TORPEDO_SSH_PASSWORD}" ]; then
    TORPEDO_SSH_PASSWORD="$TORPEDO_SSH_PASSWORD"
fi

printf "env variables:\n"
printf "VERBOSE=$VERBOSE\n"
printf "SCALE_FACTOR=$SCALE_FACTOR\n"
printf "SKIP_ARG=$SKIP_ARG\n"
printf "FOCUS_ARG=$FOCUS_ARG\n"
printf "TORPEDO_IMAGE=$TORPEDO_IMAGE\n"
printf "NOMAD_API_ADDR=$NOMAD_API_ADDR\n"
printf "TORPEDO_SSH_USER=$TORPEDO_SSH_USER\n"
printf "TORPEDO_SSH_PASSWORD=$TORPEDO_SSH_PASSWORD\n"

TORPEDO_SPEC_FILE="torpedo.nomad"
TORPEDO_JOB_NAME="torpedo-job"

printf "Building nomad torpedo yaml...\n"
cat << EOF > torpedo.nomad
job "$TORPEDO_JOB_NAME" {
  datacenters = ["dc1"]
  type = "batch"
  reschedule {
    attempts = 0
    unlimited = false
  }
  group "torpedo_group" {
    restart {
      attempts = 0
    }
    task "torpedo_task" {
      constraint {
        attribute = "\${attr.unique.network.ip-address}"
        value = "$NOMAD_API_ADDR"
      }
      driver = "docker"
      config {
        image = "$TORPEDO_IMAGE"
        args = [
          "$VERBOSE",
          "--trace",
          "--failFast",
          "$FOCUS_ARG",
          "$SKIP_ARG",
          "--slowSpecThreshold", "600",
          "bin/basic.test",
          "--",
          "--spec-dir", "../drivers/scheduler/nomad/specs",
          "--app-list", "$APP_LIST",
          "--scheduler", "nomad",
          "--node-driver", "ssh",
          "--scale-factor", "$SCALE_FACTOR"
        ]
        volumes = [
          "/var/run/docker.sock:/var/run/docker.sock",
          "/tmp/:/testresults/",
        ]
      }
      env {
        "NOMAD_API_ADDR" = "$NOMAD_API_ADDR"
        "TORPEDO_SSH_USER" = "$TORPEDO_SSH_USER"
        "TORPEDO_SSH_PASSWORD" = "$TORPEDO_SSH_PASSWORD"
      }
    }
  }
}
EOF
printf "Done building torpedo nomad spec\n"

printf "Deploy $TORPEDO_JOB_NAME and get job ID...\n"
jobId=`docker run -v ${PWD}/$TORPEDO_SPEC_FILE:/$TORPEDO_SPEC_FILE hendrikmaus/nomad-cli nomad run -address=http://$NOMAD_API_ADDR:4646 $TORPEDO_SPEC_FILE | grep "Allocation" | awk -F\" '{ printf $2 }'`

errorCode=`echo $?`
if [ "$errorCode" -ne "0" ]; then
   printf "ERROR: Got return code $errorCode instead of 0, exiting!"
   exit 1
fi

if [ -z "${jobId}" ]; then
    printf "ERROR: Did not get job ID, exiting!\n"
    exit 1
else
    printf "Job ID is $jobId\n"
fi

printf "Wait for job $jobId to start running...\n"
isRunningState=`docker run hendrikmaus/nomad-cli nomad job status -address=http://$NOMAD_API_ADDR:4646 $TORPEDO_JOB_NAME | grep $jobId | awk '{ printf $6 }'`

if [ -z "${isRunningState}" ]; then
    printf "ERROR: Did not get state\n"
    exit 1
fi

runningTimeout=0
while [ "$isRunningState" != "running" -a $runningTimeout -le 600 ]; do
    isRunningState=`docker run hendrikmaus/nomad-cli nomad job status -address=http://$NOMAD_API_ADDR:4646 $TORPEDO_JOB_NAME | grep $jobId | awk '{ printf $6 }'`
    sleep 1
    runningTimeout=$[runningTimeout+1]
    if [ "$isRunningState" == "failed" ]; then
        printf "ERROR: Torpedo job failed, exiting!\n"
        break
    fi 
done

if [ "$isCompleteState" == "failed" ]; then
    printf "Failed to start torpedo job, exiting!\n"
    exit 1
fi

if [ $runningTimeout -gt 600 ]; then
    printf "Job is still in $isRunningState..\n"
    printf "Nomad took too long to launch torpedo job. Operation timeout.\n"
    exit 1
else
    printf "Job $jobId is running\n"
fi

printf "Wait for $TORPEDO_JOB_NAME job to finish\n"
completeTimeout=0
isCompleteState=`docker run hendrikmaus/nomad-cli nomad job status -address=http://$NOMAD_API_ADDR:4646 $TORPEDO_JOB_NAME | grep $jobId | awk '{ printf $6 }'`

if [ -z "${isCompleteState}" ]; then
    printf "ERROR: Did not get state\n"
    exit 1
fi

for i in $(seq 1 600); do
    isCompleteState=`docker run hendrikmaus/nomad-cli nomad job status -address=http://$NOMAD_API_ADDR:4646 $TORPEDO_JOB_NAME | grep $jobId | awk '{ printf $6 }'`
    if [ "$isCompleteState" == "failed" ]; then
        printf "ERROR: Torpedo job failed, exiting!\n"
        break
    elif [ "$isCompleteState" == "running" ]; then
        printf "Torpedo job is still running\n"
    elif [ "$isCompleteState" == "complete" ]; then
        printf "Torpedo job successfully finished\n"
        break
    fi
    sleep 10
done

if [ "$isCompleteState" == "running" ]; then
    printf "ERROR: Took too long to complete torpedo job. Operation timeout\n"
    exit 1
fi

printf "Job $jobId is finished with status $isCompleteState\n"

printf "Print logs\n"
docker run hendrikmaus/nomad-cli nomad alloc logs -address=http://$NOMAD_API_ADDR:4646 $jobId
printf "Done\n"

if [ "$isCompleteState" == "complete" ]; then
    printf "Delete torpedo job and cleanup gc\n"
    docker run hendrikmaus/nomad-cli nomad job stop -purge -address=http://$NOMAD_API_ADDR:4646 $TORPEDO_JOB_NAME && curl -X PUT http://$NOMAD_API_ADDR:4646/v1/system/gc
    exit 0
elif [ "$isCompleteState" == "failed" ]; then
    printf "Not cleaning up failed torpedo job, exiting!\n"
    exit 1
fi
