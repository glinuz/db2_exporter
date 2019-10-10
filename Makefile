VERSION := 0.0.1
LDFLAGS := -X main.Version=$(VERSION)
GOFLAGS := -ldflags "$(LDFLAGS) -s -w"
GOARCH ?= $(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m)))
DB2HOME=/home/db2inst1/sqllib
LD_LIBRARY_PATH=$DB2HOME/lib
CGO_LDFLAGS=-L$DB2HOME/lib
CGO_CFLAGS=-I$DB2HOME/include

linux:
	@echo build linux
	@mkdir -p ./dist/ibmdb2_exporter.$(VERSION).linux-${GOARCH}
	@PKG_CONFIG_PATH=${PWD} GOOS=linux go build $(GOFLAGS) -o ./dist/ibmdb2_exporter.$(VERSION).linux-${GOARCH}/ibmdb2_exporter
	@cp default-metrics.toml ./dist/ibmdb2_exporter.$(VERSION).linux-${GOARCH}
	@(cd dist ; tar cfz ibmdb2_exporter.$(VERSION).linux-${GOARCH}.tar.gz ibmdb2_exporter.$(VERSION).linux-${GOARCH})

local-build:  linux

deps:
	@PKG_CONFIG_PATH=${PWD} go get

test:
	@echo test
	@PKG_CONFIG_PATH=${PWD} go test $$(go list ./... | grep -v /vendor/)

clean:
	@rm -rf ./dist