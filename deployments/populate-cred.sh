#!/bin/bash
set +xe
for ARGUMENT in "$@"
do
   KEY=$(echo $ARGUMENT | cut -f1 -d=)
   KEY_LENGTH=${#KEY}
   VALUE="${ARGUMENT:$KEY_LENGTH+1}"
   export "$KEY"="$VALUE"
done

sed -i 's/AWS_ACCESS_KEY_ID/${env.PSB_AWS_ACCESS_KEY_ID_PSW}/g' $filePath
sed -i 's/AWS_SECRET_ACCESS_KEY/${env.SAN_CRED_PSW}/g' $filePath
sed -i 's/AWS_REGION/$AWS_REGION/g' $filePath
sed -i 's/AZURE_ACCOUNT_NAME/${env.SAN_CRED_PSW}/g' $filePath
sed -i 's/AZURE_ACCOUNT_KEY/${env.PSB_AZURE_ACCOUNT_KEY_PSW}/g' $filePath
sed -i 's/AZURE_SUBSCRIPTION_ID/${env.AZURE_SUBSCRIPTION_ID_PSW}/g' $filePath
sed -i 's/AZURE_CLIENT_ID/${env.AZURE_CLIENT_ID_PSW}/g' $filePath
sed -i 's/AZURE_CLIENT_SECRET/${env.PSB_AZURE_CLIENT_SECRET_PSW}/g' $filePath
sed -i 's/AZURE_TENANT_ID/${env.AZURE_TENANT_ID_PSW}/g' $filePath
sed -i 's/GCP_PROJECT_ID/${env.PSB_GCP_PROJECT_ID_PSW}/g' $filePath
sed -i 's/GKE_CLUSTER_NAME/$GKE_CLUSTER_NAME/g' $filePath
sed -i 's/GKE_CLUSTER_LOCATION/$GKE_CLUSTER_LOCATION/g' $filePath
sed -i 's/GKE_PATH_TO_SERVICE_ACCOUNT_JSON/$GKE_PATH_TO_SERVICE_ACCOUNT_JSON/g' $filePath
sed -i 's/IBM_CLOUD_API_KEY/${env.PSB_IBM_CLOUD_API_KEY_PSW}/g' $filePath
sed -i 's/IBM_CLOUD_ACCOUNT_NAME/${env.PSB_IBM_CLOUD_ACCOUNT_NAME_PSW}/g' $filePath
sed -i 's/IBM_CLOUD_REGION/$IBM_CLOUD_REGION/g' $filePath
sed -i 's/IBM_CLOUD_RESOURCE_GROUP/$IBM_CLOUD_RESOURCE_GROUP/g' $filePath
sed -i 's/RANCHER_USERNAME/${env.PSB_RANCHER_USERNAME_USR}/g' $filePath
sed -i 's/RANCHER_PASSWORD/${env.PSB_RANCHER_PASSWORD_PSW}/g' $filePath
sed -i 's/CLOUD_BUCKET_NAME/$CLOUD_BUCKET_NAME/g' $filePath
sed -i 's/S3_ACCESS_KEY_ID/env.PSB_S3_ACCESS_KEY_ID_PSW/g' $filePath
sed -i 's/S3_SECRET_ACCESS_KEY/env.PSB_S3_SECRET_ACCESS_KEY_PSW/g' $filePath
sed -i 's/S3_REGION/$S3_REGION/g' $filePath
sed -i 's/SERVER_NAME/$SERVER_NAME/g' $filePath
sed -i 's/SERVER_IP/env.PSB_SERVER_IP_USR/g' $filePath
sed -i 's/EXPORT_PATH/$EXPORT_PATH/g' $filePath
sed -i 's/SUB_PATH/$SUB_PATH/g' $filePath
sed -i 's/MOUNT_OPTIONS/$MOUNT_OPTIONS/g' $filePath
sed -i 's/ENCRYPTION_KEY/$ENCRYPTION_KEY/g' $filePath