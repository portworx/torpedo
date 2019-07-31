#!/bin/sh

until psql -h ${PG_HOST} -U ${PG_USER} -c "create database ${PG_DB}" 
do
      echo "failed to create database. retrying..."
      sleep 2
done

while :
do
      pgbench -U ${PG_USER} -i -s ${TABLE_COUNT} {PG_DB};
      sleep 2

done || exit 1
