FROM golang:1.21.6-alpine

WORKDIR /go/src/github.com/portworx/

# Install setup dependencies
RUN apk update && \
    apk add --no-cache bash git gcc musl-dev make curl openssh-client ca-certificates jq libc6-compat docker python3 py3-pip

# Install Ginkgo
RUN GOFLAGS= GO111MODULE=on go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@v2.15.0

# Install aws-iam-authenticator
# This is needed for tests running inside the EKS cluster and for creating AWS entities like buckets
RUN mkdir -p /usr/local/bin && \
    curl -o /usr/local/bin/aws-iam-authenticator https://amazon-eks.s3.us-west-2.amazonaws.com/1.16.8/2020-04-16/bin/linux/amd64/aws-iam-authenticator && \
    chmod a+x /usr/local/bin/aws-iam-authenticator

# Install IBM Cloud SDK and plugins
RUN curl -fsSL https://clis.cloud.ibm.com/install/linux | sh && \
    ibmcloud plugin install -f vpc-infrastructure && \
    ibmcloud plugin install -f container-service

# Install vCluster binary
RUN curl -L -o /usr/local/bin/vcluster "https://github.com/loft-sh/vcluster/releases/latest/download/vcluster-linux-amd64" && \
    chmod 0755 /usr/local/bin/vcluster

## Install Azure CLI and dependencies
#RUN apk add --no-cache --update --virtual=build gcc musl-dev python3-dev libffi-dev openssl-dev cargo make && \
#    pip3 install "pyyaml<=5.3.1" && \
#    pip3 install --no-cache-dir --prefer-binary azure-cli && \
#    apk del build

# Install Azure CLI and dependencies within a virtual environment
RUN apk add --no-cache --update python3-dev libffi-dev openssl-dev cargo make && \
    python3 -m venv /azure-cli-venv && \
    source /azure-cli-venv/bin/activate && \
    pip3 install --upgrade pip && \
    pip3 install "pyyaml<=5.3.1" && \
    pip3 install --no-cache-dir --prefer-binary azure-cli && \
    deactivate

# Install Postman-Newman Dependencies
RUN apk update && apk upgrade \
    && apk add --no-cache \
        nodejs \
        npm \
    && rm -rf /var/cache/apk/*

# Install Newman globally using npm
RUN npm install -g newman

# Install kubectl and helm from Docker Hub
COPY --from=lachlanevenson/k8s-kubectl:latest /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=alpine/helm:latest /usr/bin/helm /usr/local/bin/helm

# Install dependancy for OCP 4.14 CLI
RUN apk --update add gcompat

# Install openssh and sshpass
RUN apk add --no-cache openssh sshpass

# Clone and build the Torpedo repository
RUN git clone "https://github.com/portworx/torpedo.git" && \
    cd torpedo && go mod download && make build build-backup