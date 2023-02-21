FROM golang:1.18.3-alpine
RUN apk update && apk add bash vim gettext jq curl

RUN chmod 776 /bin
RUN mkdir /config
RUN chmod 776 /config
RUN chmod 776 /var
RUN adduser -D consul -G root
USER consul:root

ARG consul_version="1.14.0"
ARG hashicorp_releases=https://releases.hashicorp.com
ARG PATH_BIN=/bin

# install consul
RUN wget -q -O /tmp/consul_${consul_version}_linux_amd64.zip ${hashicorp_releases}/consul/${consul_version}/consul_${consul_version}_linux_amd64.zip
RUN unzip /tmp/consul_${consul_version}_linux_amd64.zip -d ${PATH_BIN}
RUN rm -rf /tmp/consul_${consul_version}_linux_amd64.zip

ADD ./entrypoint.sh /bin/entrypoint.sh
ADD ./agent.json /tmp/agent.json
ADD ./kv-workload.sh /bin/kv-workload.sh

ENTRYPOINT ["/bin/entrypoint.sh"]
