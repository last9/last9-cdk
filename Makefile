.PHONY: test

.ONESHELL:
SHELL = /bin/bash

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

OS := linux
ifeq ($(UNAME_S),Linux)
	OS := linux
endif
ifeq ($(UNAME_S),Darwin)
	OS := darwin
endif

ARCH := amd64
ifeq ($(UNAME_M), arm64)
	ARCH := arm64
endif

lint:
	- (cd go && bash static_check.sh)

format:
	gofmt -s -w .
	goimports -w -l .

mod:
	go mod download

gotest:
	(cd go/httpmetrics && go test -v ./...)
	(cd go/sqlmetrics && env LAST9_SQL_DSN="postgres://postgres:password@127.0.0.1:8432/last9?sslmode=disable" go test -v ./...)

testup:
	docker-compose up -d wait_postgres

testdown:
	docker-compose stop -t 1
	docker-compose rm -f

test: lint testup gotest testdown
