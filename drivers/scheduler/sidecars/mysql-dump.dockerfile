FROM mysql

COPY www/WordPress.sql .
COPY scripts/mysql-dump.sh .

CMD ./mysql-dump.sh
