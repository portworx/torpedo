FROM wordpress:cli

WORKDIR /home/www-data

USER root
RUN apk add --update bash less curl
USER www-data

COPY scripts/wp-install.sh .

CMD bash ./wp-install.sh
