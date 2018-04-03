FROM alpine

WORKDIR /src
COPY www .

CMD cp -r html/* /wordpress/
