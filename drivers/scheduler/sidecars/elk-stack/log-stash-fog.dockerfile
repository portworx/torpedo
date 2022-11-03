#FROM alpine/git AS downloader
FROM golang:1.17.13

ENV CGO_ENABLED=0
ENV GO111MODULE=on

USER root
WORKDIR /go/src
RUN cd /go/src && git clone https://github.com/mingrammer/flog.git


#COPY --from=downloader /go/src/flog/go.mod ./
#COPY --from=downloader /go/src/flog/go.sum ./
#RUN go mod download
#COPY . ./
#RUN ls -alt /go/src/flog

#RUN mkdir /bin
RUN cd /go/src/flog && go build -o /bin/flog

#FROM scratch
#COPY --from=0 /bin/flog /bin/flog
COPY scripts/elk-stack/flog.sh /go/src/entry-point.sh
RUN chmod 777 /go/src/entry-point.sh
ENTRYPOINT ["/go/src/entry-point.sh"]
#ENTRYPOINT ["/bin/sh /go/src/entry-point.sh"]
