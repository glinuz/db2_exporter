# IBM DB2 exporter 

A Prometheus exporter for IBM DB2

## Description
The following metrics are exposed currently.

    ibmdb2_exporter_last_scrape_duration_seconds
    ibmdb2_exporter_last_scrape_error
    ibmdb2_exporter_scrapes_total
    ibmdb2_up
    ibmdb2_tablespace_total_mb
    ibmdb2_tablespace_free_mb
    ibmdb2_tablespace_uesd_mb

## Build

Use DB2 instance,db2inst1, in your database server HOME.
```
IBM_DB_DIR=/home/db2inst1/sqllib
export CGO_LDFLAGS=-L$IBM_DB_DIR/lib
export CGO_CFLAGS=-I$IBM_DB_DIR/include
go build main.go
```

## Runing

Maybe need set export LD_LIBRARY_PATH ,check your system env
```
export LD_LIBRARY_PATH=$DB2_HOME/lib:$LD_LIBRARY_PATH
```
Database connect setting as env DB2_DSN,default DATABASE=sample; HOSTNAME=localhost; PORT=60000; PROTOCOL=TCPIP; UID=db2inst1; PWD=db2inst1;

Running on port 9161
```
export DB2_DSN="DATABASE=sample; HOSTNAME=localhost; PORT=60000; PROTOCOL=TCPIP; UID=db2inst1; PWD=db2inst1;"
./main 
```