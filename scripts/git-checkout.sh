#!/bin/sh
set -ex

cd /github.com/clone

if [ -z "$BRANCH" ]; then
  echo "Error: Unable to check out because the BRANCH environment variable is not set"
  exit 1
fi

ls

git status 2>/dev/null
git --no-pager log -1 --oneline

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [ "$CURRENT_BRANCH" = "$BRANCH" ]; then
  git fetch origin ${BRANCH} --quiet
  git reset --hard origin/${BRANCH}
else
  git fetch origin ${BRANCH}:${BRANCH} --quiet
  git checkout ${BRANCH}
fi

git --no-pager log -1 --oneline

echo "Time Start: [$(date +'%Y-%m-%d %H:%M:%S')]"

exec "$@"
