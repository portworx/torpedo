#!/bin/sh

echo "Using host: ${PG_HOST} user: ${PG_USER}"
until psql -h ${PG_HOST} -U ${PG_USER} -c "create database pxdemo"
do
      echo "failed to create database. retrying..."
      sleep 2
done

while :
do
      pgbench -h ${PG_HOST} -U ${PG_USER} -i -s ${SIZE} pxdemo;
      sleep 2

done || exit 1
