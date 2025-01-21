package hammerdb

import (
	"github.com/pure-px/sched-ops/k8s/core"
)

var (
	k8sCore = core.Instance()
)

var tclScriptMssqlTprocch = "puts \"SETTING CONFIGURATION\"\n\n" + "dbset db mssqls\n" + "dbset bm TPC-H\n" + "diset connection mssqls_tcp false\n" +
	"diset connection mssqls_port 1433\n" + "diset connection mssqls_azure false\n" +
	"diset connection mssqls_encrypt_connection true\n" + "diset connection mssqls_trust_server_cert true\n" +
	"diset connection mssqls_authentication sql\n" + "diset connection mssqls_server %s\n" +
	"diset connection mssqls_linux_server %s\n" + "diset connection mssqls_uid %s\n" +
	"diset connection mssqls_pass %s\n" + "diset connection mssqls_linux_authent sql\n" +
	"diset connection mssqls_linux_odbc {ODBC Driver 18 for SQL Server}\n\n" + "diset tpch mssqls_scale_fact %d\n" +
	"diset tpch mssqls_maxdop 2\n" + "diset tpch mssqls_num_tpch_threads [ numberOfCPUs ]\n" +
	"diset tpch mssqls_tpch_dbase %s\n" + "diset tpch mssqls_colstore false\n" + "diset tpch mssqls_tpch_use_bcp false\n\n" + "puts \"SCHEMA BUILD STARTED\"\n" +
	"buildschema\n" + "puts \"SCHEMA BUILD COMPLETED\""

var triggerScript = "export LD_LIBRARY_PATH=/home/instantclient_21_5/::/usr/local/unixODBC/lib" + "\n" +
	"export ORACLE_LIBRARY=/home/instantclient_21_5/libclntsh.so" + "export ODBCINI=/usr/local/unixODBC/etc/odbc.ini" + "\n" +
	"export ODBCSYSINI=/usr/local/unixODBC/etc\n" + "export TMP=`pwd`/TMP\n" + "mkdir -p $TMP" + "\n\n" +
	"%shammerdbcli auto %s%s"
