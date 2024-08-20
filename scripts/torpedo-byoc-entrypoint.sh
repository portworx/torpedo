#!/bin/bash
set -e

# Prints timestamp before executing the command
timestamped_exec() {
  echo "$(date +'%Y-%m-%d %H:%M:%S') - Executing: $*"
  "$@"
}

# The default entrypoint is executed if the IS_BYOC_PIPELINE env variable is not set to true
if [[ "$IS_BYOC_PIPELINE" != "true" && "$IS_BYOC_PIPELINE" != true ]]; then
  timestamped_exec ginkgo --fail-fast --poll-progress-after 3m -v -trace --junit-report=/testresults/junit_basic.xml
  exit $?
fi

# It is assumed that the Torpedo repository is already cloned at /github.com/clone/torpedo
timestamped_exec cd /github.com/clone/torpedo

# Validate Torpedo branch
if [ -z "$BRANCH" ]; then
  echo "$(date +'%Y-%m-%d %H:%M:%S') - Error: Unable to check out because the BRANCH env variable is not set"
  exit 1
fi

# Prints the current branch and commit hash
timestamped_exec git status 2>/dev/null
timestamped_exec git --no-pager log -1 --oneline

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# If the current branch is the same as the desired branch, fetches the latest changes
if [ "$CURRENT_BRANCH" = "$BRANCH" ]; then
  timestamped_exec git fetch origin ${BRANCH} --quiet
  timestamped_exec git reset --hard origin/${BRANCH}
# Else it fetches the desired branch from the remote and checks it out
else
  timestamped_exec git fetch origin ${BRANCH}:${BRANCH} --quiet
  timestamped_exec git checkout ${BRANCH}
fi

# Prints the current branch and commit hash
timestamped_exec git status 2>/dev/null
timestamped_exec git --no-pager log -1 --oneline

# Runs pre-torpedo script before running Ginkgo command
if [[ "$RUN_PRE_TORPEDO_SCRIPT" != "true" && "$RUN_PRE_TORPEDO_SCRIPT" != true ]]; then
  timestamped_exec chmod +x /scripts/pre-torpedo-script.sh
  timestamped_exec /scripts/pre-torpedo-script.sh
  exit $?
fi

# Runs Ginkgo command
timestamped_exec ginkgo "$@"

# Runs post-torpedo script after running Ginkgo command
if [[ "$RUN_POST_TORPEDO_SCRIPT" != "true" && "$RUN_POST_TORPEDO_SCRIPT" != true ]]; then
  timestamped_exec chmod +x /scripts/post-torpedo-script.sh
  timestamped_exec /scripts/post-torpedo-script.sh
  exit $?
fi