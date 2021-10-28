#!/bin/sh
workload=1

oltp() {
  while :
  do
    sysbench --db-driver=mysql --oltp-table-size=2000000 --oltp-tables-count=64 --threads=64 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/parallel_prepare.lua prepare
    sysbench --db-driver=mysql --report-interval=1 --oltp-table-size=2000000 --threads=64 --delete_inserts=10 --index_updates=10 --non_index_updates=10 --db-ps-mode=disable --time=3000 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/oltp.lua run
    sysbench --db-driver=mysql --oltp-table-size=2000000 --oltp-tables-count=64 --threads=64 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/oltp.lua cleanup
    sleep 2
  done
}

basic() {
  while :
  do
    sysbench --db-driver=mysql --oltp-table-size=10000 --oltp-tables-count=10 --threads=1 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/parallel_prepare.lua prepare
    sysbench --db-driver=mysql --report-interval=1 --mysql-table-engine=innodb --oltp-table-size=10000 --threads=1 --time=5400 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password --olpt-skip-trx /usr/share/sysbench/tests/include/oltp_legacy/oltp.lua run
    sysbench --db-driver=mysql --oltp-table-size=10000 --oltp-tables-count=10 --threads=1 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/oltp.lua cleanup
    sleep 2
  done
}

exit_script() {
    if [ $workload = 2 ];
    then
      sysbench --db-driver=mysql --oltp-table-size=2000000 --oltp-tables-count=64 --threads=64 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/oltp.lua cleanup
    else
      sysbench --db-driver=mysql --oltp-table-size=10000 --oltp-tables-count=10 --threads=1 --mysql-host=${MYSQL_HOST} --mysql-port=3306 --mysql-user=sysbench --mysql-password=password /usr/share/sysbench/tests/include/oltp_legacy/oltp.lua cleanup
    fi
    trap - INT TERM # clear the trap
    kill -- -$$ # Sends SIGTERM to child/sub processes
}

trap exit_script INT TERM

while [ "$1" != "" ]; do
  case $1 in
    --oltp )  shift
              workload=2
              ;;
    * ) exit 1
  esac
done

until mysql -h ${MYSQL_HOST} -u root -p${MYSQL_ROOT_PASSWORD} -e "CREATE SCHEMA IF NOT EXISTS sbtest ; CREATE USER IF NOT EXISTS sysbench@'%' IDENTIFIED BY 'password'; GRANT ALL PRIVILEGES ON sbtest.* to sysbench@'%';"
do
  echo "failed to create schema and user. retrying..."
  sleep 2
done

if [ $workload = 2 ];
then
  echo "running sysbench oltp workload"
  oltp
  exit 0
fi

echo "running sysbench basic workload"
basic