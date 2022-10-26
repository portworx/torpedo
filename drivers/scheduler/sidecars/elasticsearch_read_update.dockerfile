FROM alpine/git AS downloader
USER root
RUN cd /tmp && git clone https://github.com/logzio/elasticsearch-stress-test.git
FROM python:2
WORKDIR /usr/src/elasticsearch-stress-test
COPY --from=downloader /tmp/elasticsearch-stress-test/elasticsearch-stress-test.py ./
COPY --from=downloader /tmp/elasticsearch-stress-test/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY scripts/elasticsearch/elasticsearch_readupdate.py ./
COPY scripts/elasticsearch/esreadupdate.sh ./
RUN chmod 777 ./esreadupdate.sh
ENTRYPOINT ["sh", "/esreadupdate.sh"]
