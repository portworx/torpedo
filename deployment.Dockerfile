# This Dockerfile sets up the environment for deploying Torpedo as a Kubernetes pod.

FROM golang:1.21.6-alpine

# The PATH incldues /go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apk add --no-cache bash

# The set -o pipefail directive is used with /bin/bash -c interpreter
# to ensure that any error at any stage of the pipe fails the command, preventing inadvertent build success.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# kubectl
COPY --from=bitnami/kubectl:1.29.1 /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/

ARG GINKGO_CLI_VERSION="v2.15.0"

# The default command starts a bash shell.
# Users can override this command when running the container to execute specific scripts or commands.
CMD ["/bin/bash"]
