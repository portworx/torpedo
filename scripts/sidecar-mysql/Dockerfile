FROM alpine

RUN apk add --no-cache --virtual .build-deps curl bzip2 tar 
RUN curl -L https://launchpad.net/test-db/employees-db-1/1.0.6/+download/employees_db-full-1.0.6.tar.bz2 | bzip2 -dfv - | tar -xf -
RUN apk del .build-deps
RUN apk add --no-cache mysql-client
COPY entry-point.sh .
RUN chmod u+x entry-point.sh
ENTRYPOINT ["/entry-point.sh"]