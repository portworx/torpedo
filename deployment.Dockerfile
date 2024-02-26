# This Dockerfile sets up the environment for deploying Torpedo as a Kubernetes pod.

FROM golang:1.21.6-alpine

# The PATH incldues /go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Install bash and git
RUN apk add --no-cache bash git

# The set -o pipefail directive is used with /bin/bash -c interpreter
# to ensure that any error at any stage of the pipe fails the command, preventing inadvertent build success.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# kubectl
COPY --from=bitnami/kubectl:1.29.1 /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/


#WORKDIR /go/src/github.com/torpedo

CMD ["/bin/bash"]
