#!/bin/sh
trap '' INT
trap '' QUIT
trap '' STOP

if [ -z "${SIZE}" ]; then
    echo "SIZE is not defined, exiting..."
    exit 1
fi

until mysql -h ${MYSQL_HOST} -u root -p${MYSQL_ROOT_PASSWORD} -e "CREATE SCHEMA IF NOT EXISTS sbtest ; CREATE USER IF NOT EXISTS sysbench@'%' IDENTIFIED BY 'password'; GRANT ALL PRIVILEGES ON sbtest.* to sysbench@'%';"
do
  echo "failed to create schema and user. retrying..."
  sleep 2
done
while :
do
    sysbench --db-driver=mysql --oltp-table-size=10000 --oltp-tables-count=10 --threads=1 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/parallel_prepare.lua prepare
    sysbench --db-driver=mysql --report-interval=1 --mysql-table-engine=innodb --oltp-table-size=10000 --threads=1 --time=5400 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password --olpt-skip-trx /usr/share/sysbench/tests/include/oltp_legacy/oltp.lua run

    expected_data_size=$(($SIZE * 1024 * 1024))
    actual_data_size=$(du -s /var/lib/mysql/* | cut -f1)

    if [ $actual_data_size -le $expected_data_size ]; then
      sleep 2
    else
      break
    fi
done

sysbench --db-driver=mysql --oltp-table-size=10000 --oltp-tables-count=10 --threads=1 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/oltp.lua cleanup
