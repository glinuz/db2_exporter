# IBM DB2 exporter

A Prometheus exporter for IBM DB2,with getting by ODBC.

## Description

The following metrics are exposed currently by default-metrics.toml.

```html
    # HELP ibmdb2_bufferpool_bp_size_mb DB2 bufferpools size MB.
    # TYPE ibmdb2_bufferpool_bp_size_mb gauge
    ibmdb2_bufferpool_bp_size_mb{bp_name="BUFF16              "} 300
    ibmdb2_bufferpool_bp_size_mb{bp_name="IBMDEFAULTBP        "} 800
    # HELP ibmdb2_bufferpool_idx_hit_ratio DB2 bufferpools INDEX_HIT_RATIO_PERCENT.
    # TYPE ibmdb2_bufferpool_idx_hit_ratio gauge
    ibmdb2_bufferpool_idx_hit_ratio{bp_name="IBMDEFAULTBP        "} 100
    # HELP ibmdb2_bufferpool_total_hit_ratio DB2 bufferpools TOTAL_HIT_RATIO_PERCENT.
    # TYPE ibmdb2_bufferpool_total_hit_ratio gauge
    ibmdb2_bufferpool_total_hit_ratio{bp_name="BUFF16              "} 100
    ibmdb2_bufferpool_total_hit_ratio{bp_name="IBMDEFAULTBP        "} 100
    # HELP ibmdb2_db_name_value DB2 name.
    # TYPE ibmdb2_db_name_value gauge
    ibmdb2_db_name_value{server="SAMPLE"} 1
    # HELP ibmdb2_db_version_value DB2 version.
    # TYPE ibmdb2_db_version_value gauge
    ibmdb2_db_version_value{service_level="DB2 v10.5.0.6"} 1
    # HELP ibmdb2_exporter_last_scrape_duration_seconds Duration of the last scrape of metrics from IBM DB2.
    # TYPE ibmdb2_exporter_last_scrape_duration_seconds gauge
    ibmdb2_exporter_last_scrape_duration_seconds 0.013067445
    # HELP ibmdb2_exporter_last_scrape_error Whether the last scrape of metrics from IBM DB2 resulted in an error (1 for error, 0 for success).
    # TYPE ibmdb2_exporter_last_scrape_error gauge
    ibmdb2_exporter_last_scrape_error 0
    # HELP ibmdb2_exporter_scrapes_total Total number of times IBM DB2 was scraped for metrics.
    # TYPE ibmdb2_exporter_scrapes_total counter
    ibmdb2_exporter_scrapes_total 2
    # HELP ibmdb2_tablespace_free_mb Tablespaces free space MB.
    # TYPE ibmdb2_tablespace_free_mb gauge
    ibmdb2_tablespace_free_mb{page_size="8192",state="NORMAL    ",tablespace="IBMDB2SAMPLEREL     ",type="LARGE     "} 27
    ibmdb2_tablespace_free_mb{page_size="8192",state="NORMAL    ",tablespace="IBMDB2SAMPLEXML     ",type="LARGE     "} 20
    ibmdb2_tablespace_free_mb{page_size="8192",state="NORMAL    ",tablespace="SYSCATSPACE         ",type="ANY       "} 4
    ibmdb2_tablespace_free_mb{page_size="8192",state="NORMAL    ",tablespace="SYSTOOLSPACE        ",type="LARGE     "} 31
    ibmdb2_tablespace_free_mb{page_size="8192",state="NORMAL    ",tablespace="TEMPSPACE1          ",type="SYSTEMP   "} 0
    ibmdb2_tablespace_free_mb{page_size="8192",state="NORMAL    ",tablespace="USERSPACE1          ",type="LARGE     "} 17
    # HELP ibmdb2_tablespace_total_mb Tablespaces total space MB.
    # TYPE ibmdb2_tablespace_total_mb gauge
    ibmdb2_tablespace_total_mb{page_size="8192",state="NORMAL    ",tablespace="IBMDB2SAMPLEREL     ",type="LARGE     "} 32
    ibmdb2_tablespace_total_mb{page_size="8192",state="NORMAL    ",tablespace="IBMDB2SAMPLEXML     ",type="LARGE     "} 32
    ibmdb2_tablespace_total_mb{page_size="8192",state="NORMAL    ",tablespace="SYSCATSPACE         ",type="ANY       "} 128
    ibmdb2_tablespace_total_mb{page_size="8192",state="NORMAL    ",tablespace="SYSTOOLSPACE        ",type="LARGE     "} 32
    ibmdb2_tablespace_total_mb{page_size="8192",state="NORMAL    ",tablespace="TEMPSPACE1          ",type="SYSTEMP   "} 0
    ibmdb2_tablespace_total_mb{page_size="8192",state="NORMAL    ",tablespace="USERSPACE1          ",type="LARGE     "} 32
    # HELP ibmdb2_tablespace_used_mb Tablespaces used space MB.
    # TYPE ibmdb2_tablespace_used_mb gauge
    ibmdb2_tablespace_used_mb{page_size="8192",state="NORMAL    ",tablespace="IBMDB2SAMPLEREL     ",type="LARGE     "} 4
    ibmdb2_tablespace_used_mb{page_size="8192",state="NORMAL    ",tablespace="IBMDB2SAMPLEXML     ",type="LARGE     "} 11
    ibmdb2_tablespace_used_mb{page_size="8192",state="NORMAL    ",tablespace="SYSCATSPACE         ",type="ANY       "} 123
    ibmdb2_tablespace_used_mb{page_size="8192",state="NORMAL    ",tablespace="SYSTOOLSPACE        ",type="LARGE     "} 0
    ibmdb2_tablespace_used_mb{page_size="8192",state="NORMAL    ",tablespace="TEMPSPACE1          ",type="SYSTEMP   "} 0
    ibmdb2_tablespace_used_mb{page_size="8192",state="NORMAL    ",tablespace="USERSPACE1          ",type="LARGE     "} 14
    # HELP ibmdb2_up Whether the IBM DB2 database server is up.
    # TYPE ibmdb2_up gauge
    ibmdb2_up 1
```
## Linux unix macOS
### Build (go <1.11 vendor mode)

Use DB2 server or client instance ,like db2inst1, in your database server HOME directory.
Need DB2 ODBC driver file.



```bash
export IBM_DB_DIR=/home/db2inst1/sqllib
export CGO_LDFLAGS=-L$IBM_DB_DIR/lib
export CGO_CFLAGS=-I$IBM_DB_DIR/include
go build main.go
```

### Run

Switch DB2 server CFG monitor buffpool on

```bash
db2 update monitor switches using bufferpool on
db2 update dbm cfg using DFT_MON_BUFPOOL on
```

Set export LD_LIBRARY_PATH ,check your system ENV.

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
## Windows 
see issue

## Zabbix template

In our case,it is worked with Prometheus and Zabbix (v3.2).
Import db2export_zabbix_templates.xml, and define Host macro {$URL} endpoint,e.g. <http://localhost:9161/metrics>

## Howto custerm metric

Add your custerm metric ,database monitor SQL script in file: default-metrics.toml.



![DB2 export](https://github.com/glinuz/ibmdb2_exporter/blob/master/ibmdb2.png)
