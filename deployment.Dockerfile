# This Dockerfile sets up the environment for deploying Torpedo.

# Alpine Linux is chosen as the base image for its security, simplicity and resource efficiency.
# For release notes, visit: https://alpinelinux.org/releases/
# It uses its own package manager called apk
FROM alpine:3.19.1

# Define arguments for versions of tools to be installed.
# These versions can be updated as new releases become available.
# Reference: https://github.com/kubernetes/kubernetes/releases/tag/v1.29.1
ARG KUBECTL_VERSION="v1.29.1"
# Reference: https://golang.org/dl/#go1.21.6
ARG GO_VERSION="1.21.6"
# Reference: https://github.com/onsi/ginkgo/releases/tag/v2.15.0
ARG GINKGO_VERSION="v2.15.0"

# Set the default shell to bash with pipefail option for better error handling in shell commands.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install runtime dependencies required for Torpedo and its deployment scripts.
RUN apk add --no-cache \
    bash \
    curl \
    git \
    jq

# Install build dependencies necessary for compiling any dependencies or tools from source.
# These are temporary and will be removed to keep the final image size down.
RUN apk add --no-cache --virtual .build-deps \
    gcc \
    musl-dev

# Download and install Go programming language.
# Go is required for building and running Go-based applications, including Torpedo and Ginkgo tests.
RUN wget -q "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# Install kubectl, a command-line tool for Kubernetes cluster management.
# This is required for managing Kubernetes resources as part of Torpedo's deployment and testing.
RUN curl -sLO "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/

# Install Ginkgo, a Go testing framework, for running Torpedo's test suite.
# Ginkgo provides advanced testing capabilities, such as BDD-style tests.
RUN GOPATH=$(mktemp -d) \
    && export PATH="/usr/local/go/bin:${GOPATH}/bin:${PATH}" \
    && go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION} \
    && mv ${GOPATH}/bin/ginkgo /usr/local/bin/ \
    && rm -rf ${GOPATH}

# Cleanup the build dependencies to reduce the image size.
# This step ensures that only the runtime dependencies and necessary tools remain in the final image.
RUN apk del .build-deps

# Set the working directory to /torpedo.
# This directory will be the default location for running Torpedo commands.
WORKDIR /torpedo

# Set environment variables for Go.
# These are necessary for running Go commands and for Go-based applications to locate the Go installation.
ENV GOROOT="/usr/local/go"
ENV GOPATH="/go"
ENV PATH="$PATH:$GOROOT/bin:$GOPATH/bin"

# The default command starts a bash shell.
# Users can override this command when running the container to execute specific scripts or commands.
CMD ["/bin/bash"]

# Instructions for building and running this Docker image:
# To build the image, navigate to the directory containing this Dockerfile and run:
# docker build -t torpedo-deployment:latest -f deployment.Dockerfile .
#
# To run the image interactively with a bash shell, use:
# docker run -it torpedo-deployment:latest
#
# Ensure you have Docker installed and configured on your system.
# For more detailed instructions on using Docker, refer to the Docker documentation: https://docs.docker.com/get-started/
