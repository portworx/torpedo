#!/bin/sh
set -xe
trap '' SIGINT
trap ''  SIGQUIT
trap '' SIGTSTP

cd /employees_db
until mysql -h mysql -u root -p${MYSQL_ROOT_PASSWORD} -t < employees.sql
do 
	echo "adding data"
	sleep 2
done

while :
do
	c=$(($RANDOM%6))
	case $c in
		1) mysql -h mysql -u root -p${MYSQL_ROOT_PASSWORD} -D employees -e "SELECT * FROM employees LIMIT 1000";;
		2) mysql -h mysql -u root -p${MYSQL_ROOT_PASSWORD} -D employees -e "SELECT * FROM titles LIMIT 1000";;
		3) mysql -h mysql -u root -p${MYSQL_ROOT_PASSWORD} -D employees -e "SELECT * FROM dept_emp LIMIT 1000";;
		4) mysql -h mysql -u root -p${MYSQL_ROOT_PASSWORD} -D employees -e "SELECT * FROM dept_manager LIMIT 1000";;
		5) mysql -h mysql -u root -p${MYSQL_ROOT_PASSWORD} -D employees -e "SELECT * FROM departments LIMIT 1000";;
		*) ;;
	esac
	sleep 2
done






