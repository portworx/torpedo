#!/bin/bash

cd ./torpedo || exit
git status
git stash push -u
git fetch origin "${BRANCH}" --quiet
git checkout "${BRANCH}"
git reset --hard origin/"${BRANCH}"
make "${CONTAINER_TARGET}"
ginkgo --trace \
  --timeout "${TIMEOUT}" \
  "${FAIL_FAST}" \
  --poll-progress-after 10m \
  --junit-report="${JUNIT_REPORT_PATH}" \
  "${FOCUS_ARG}" \
  "${SKIP_ARG}" \
  "${TEST_SUITE}" -- \
  --spec-dir "${SPEC_DIR}" \
  --app-list "${APP_LIST}" \
  --deploy-pds-apps="${DEPLOY_PDS_APPS}" \
  --pds-driver "${PDS_DRIVER}" \
  --secure-apps "${SECURE_APP_LIST}" \
  --repl1-apps "${REPL1_APP_LIST}" \
  --csi-app-list "${CSI_APP_LIST}" \
  --scheduler "${SCHEDULER}" \
  --max-storage-nodes-per-az "${MAX_STORAGE_NODES_PER_AZ}" \
  --backup-driver "${BACKUP_DRIVER}" \
  --log-level "${LOGLEVEL}" \
  --node-driver "${NODE_DRIVER}" \
  --scale-factor "${SCALE_FACTOR}" \
  --hyper-converged="${IS_HYPER_CONVERGED}" \
  --fail-on-px-pod-restartcount="${PX_POD_RESTART_CHECK}" \
  --minimun-runtime-mins "${MIN_RUN_TIME}" \
  --driver-start-timeout "${DRIVER_START_TIMEOUT}" \
  --chaos-level "${CHAOS_LEVEL}" \
  --storagenode-recovery-timeout "${STORAGENODE_RECOVERY_TIMEOUT}" \
  --provisioner "${PROVISIONER}" \
  --storage-driver "${STORAGE_DRIVER}" \
  --config-map "${CONFIGMAP}" \
  --custom-config "${CUSTOM_APP_CONFIG_PATH}" \
  --storage-upgrade-endpoint-url="${UPGRADE_ENDPOINT_URL}" \
  --storage-upgrade-endpoint-version="${UPGRADE_ENDPOINT_VERSION}" \
  --upgrade-storage-driver-endpoint-list="${UPGRADE_STORAGE_DRIVER_ENDPOINT_LIST}" \
  --enable-stork-upgrade="${ENABLE_STORK_UPGRADE}" \
  --secret-type="${SECRET_TYPE}" \
  --pure-volumes="${IS_PURE_VOLUMES}" \
  --pure-fa-snapshot-restore-to-many-test="${PURE_FA_CLONE_MANY_TEST}" \
  --pure-san-type="${PURE_SAN_TYPE}" \
  --vault-addr="${VAULT_ADDR}" \
  --vault-token="${VAULT_TOKEN}" \
  --px-runtime-opts="${PX_RUNTIME_OPTS}" \
  --px-cluster-opts="${PX_CLUSTER_OPTS}" \
  --anthos-ws-node-ip="${ANTHOS_ADMIN_WS_NODE}" \
  --anthos-inst-path="${ANTHOS_INST_PATH}" \
  --autopilot-upgrade-version="${AUTOPILOT_UPGRADE_VERSION}" \
  --csi-generic-driver-config-map="${CSI_GENERIC_CONFIGMAP}" \
  --sched-upgrade-hops="${SCHEDULER_UPGRADE_HOPS}" \
  --migration-hops="${MIGRATION_HOPS}" \
  --license_expiry_timeout_hours="${LICENSE_EXPIRY_TIMEOUT_HOURS}" \
  --metering_interval_mins="${METERING_INTERVAL_MINS}" \
  --testrail-milestone="${TESTRAIL_MILESTONE}" \
  --testrail-run-name="${TESTRAIL_RUN_NAME}" \
  --testrail-run-id="${TESTRAIL_RUN_ID}" \
  --testrail-jeknins-build-url="${TESTRAIL_JENKINS_BUILD_URL}" \
  --testrail-host="${TESTRAIL_HOST}" \
  --testrail-username="${TESTRAIL_USERNAME}" \
  --testrail-password="${TESTRAIL_PASSWORD}" \
  --jira-username="${JIRA_USERNAME}" \
  --jira-token="${JIRA_TOKEN}" \
  --jira-account-id="${JIRA_ACCOUNT_ID}" \
  --user="${USER}" \
  --enable-dash="${ENABLE_DASH}" \
  --data-integrity-validation-tests="${DATA_INTEGRITY_VALIDATION_TESTS}" \
  --test-desc="${TEST_DESCRIPTION}" \
  --test-type="${TEST_TYPE}" \
  --test-tags="${TEST_TAGS}" \
  --testset-id="${DASH_UID}" \
  --branch="${BRANCH}" \
  --product="${PRODUCT}" \
  --torpedo-job-name="${TORPEDO_JOB_NAME}" \
  --torpedo-job-type="${TORPEDO_JOB_TYPE}" \
  --torpedo-skip-system-checks="${TORPEDO_SKIP_SYSTEM_CHECKS}" \
  "${APP_DESTROY_TIMEOUT_ARG}" \
  "${SCALE_APP_TIMEOUT_ARG}"
