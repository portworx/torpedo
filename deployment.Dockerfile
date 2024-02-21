# This Dockerfile sets up the environment for deploying Torpedo as a Kubernetes pod.

# Alpine Linux is chosen as the base image for its security, simplicity, and resource efficiency.
FROM alpine:3.19.1

# The set -o pipefail directive is used with /bin/bash -c interpreter
# to ensure that any error at any stage of the pipe fails the command, preventing inadvertent build success.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Note: Define arguments to specify dependency versions,
# to facilitate building of images with different versions without modifying the Dockerfile.

# Install the Go Programming Language.
ARG GO_VERSION="1.21.6"
RUN wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz -O go.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz
# Set the environment variables required for Go.
ENV GOPATH="/go"
ENV GOROOT="/usr/local/go"
ENV PATH="$PATH:$GOROOT/bin:$GOPATH/bin"

# KUBECTL_VERSION specifies the kubectl version to be installed.
ARG KUBECTL_VERSION="v1.29.1"

# GINKGO_CLI_VERSION specifies the Ginkgo CLI version to be installed.
ARG GINKGO_CLI_VERSION="v2.15.0"

# The default command starts a bash shell.
# Users can override this command when running the container to execute specific scripts or commands.
CMD ["/bin/bash"]
