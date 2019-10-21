# IBM DB2 exporter

A Prometheus exporter for IBM DB2.

## Description

The following metrics are exposed currently by default-metrics.toml.

    # HELP ibmdb2_bufferpool_bp_size_mb DB2 bufferpools size MB.
    # TYPE ibmdb2_bufferpool_bp_size_mb gauge
    ibmdb2_bufferpool_bp_size_mb{bp_name="IBMDEFAULTBP        "} 7
    # HELP ibmdb2_db_value DB2 databaes name.
    # TYPE ibmdb2_db_value gauge
    ibmdb2_db_value{dbname="SAMPLE"} 1
    # HELP ibmdb2_tablespace_free_mb Tablespaces free space MB.
    # TYPE ibmdb2_tablespace_free_mb gauge
    ibmdb2_tablespace_free_mb{page_size="8192",tablespace="IBMDB2SAMPLEREL     ",type="LARGE     "} 27
    ibmdb2_tablespace_free_mb{page_size="8192",tablespace="IBMDB2SAMPLEXML     ",type="LARGE     "} 20
    ibmdb2_tablespace_free_mb{page_size="8192",tablespace="SYSCATSPACE         ",type="ANY       "} 4
    ibmdb2_tablespace_free_mb{page_size="8192",tablespace="SYSTOOLSPACE        ",type="LARGE     "} 31
    ibmdb2_tablespace_free_mb{page_size="8192",tablespace="TEMPSPACE1          ",type="SYSTEMP   "} 0
    ibmdb2_tablespace_free_mb{page_size="8192",tablespace="USERSPACE1          ",type="LARGE     "} 17
    # HELP ibmdb2_tablespace_total_mb Tablespaces total space MB.
    # TYPE ibmdb2_tablespace_total_mb gauge
    ibmdb2_tablespace_total_mb{page_size="8192",tablespace="IBMDB2SAMPLEREL     ",type="LARGE     "} 32
    ibmdb2_tablespace_total_mb{page_size="8192",tablespace="IBMDB2SAMPLEXML     ",type="LARGE     "} 32
    ibmdb2_tablespace_total_mb{page_size="8192",tablespace="SYSCATSPACE         ",type="ANY       "} 128
    ibmdb2_tablespace_total_mb{page_size="8192",tablespace="SYSTOOLSPACE        ",type="LARGE     "} 32
    ibmdb2_tablespace_total_mb{page_size="8192",tablespace="TEMPSPACE1          ",type="SYSTEMP   "} 0
    ibmdb2_tablespace_total_mb{page_size="8192",tablespace="USERSPACE1          ",type="LARGE     "} 32
    # HELP ibmdb2_up Whether the IBM DB2 database server is up.
    # TYPE ibmdb2_up gauge
    ibmdb2_up 1
    # HELP ibmdb2_version_value DB2 version.
    # TYPE ibmdb2_version_value gauge
    ibmdb2_version_value{service_level="DB2 v10.5.0.6"} 1

## Build

Use DB2 instance,like db2inst1, in your database server HOME directory.

```bash
IBM_DB_DIR=/home/db2inst1/sqllib
export CGO_LDFLAGS=-L$IBM_DB_DIR/lib
export CGO_CFLAGS=-I$IBM_DB_DIR/include
go build main.go
```

## Run

Maybe need set export LD_LIBRARY_PATH ,check your system ENV.

```bash
export LD_LIBRARY_PATH=$DB2_HOME/lib:$LD_LIBRARY_PATH
```

Database connect setting as env DB2_DSN,default DATABASE=sample; HOSTNAME=localhost; PORT=60000; PROTOCOL=TCPIP; UID=db2inst1; PWD=db2inst1;

Running on port 9161

```bash
export DB2_DSN="DATABASE=sample; HOSTNAME=localhost; PORT=60000; PROTOCOL=TCPIP; UID=db2inst1; PWD=db2inst1;"
./main
```
OR

```bash
./main -log.level debug -dsn "DATABASE=sample; HOSTNAME=localhost; PORT=60000; PROTOCOL=TCPIP; UID=db2inst1; PWD=db2inst1;"
```

## Zabbix template
In our case,it is worked with Prometheus or Zabbix. 
Import db2export_zabbix_templates.xml, and define Host macro {$URL} endpoint,e.g. http://localhost:9161/metrics

## Howto custerm metric

Add your custerm metric ,database monitor SQL script in file: default-metrics.toml.

![DB2 export](https://github.com/glinuz/ibmdb2_exporter/blob/master/ibmdb2.png)
