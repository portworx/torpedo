FROM golang:1.18.3-alpine
RUN apk update && apk add bash vim

# install consul-bench
WORKDIR /consul-bench
ADD . /consul-bench
RUN go build -o /bin/consul-bench && rm -rf /consul-bench
WORKDIR /

ADD ./entrypoint.sh /bin/entrypoint.sh
ENTRYPOINT ["/bin/entrypoint.sh"]
