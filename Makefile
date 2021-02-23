APPVERSION := $(shell cat ./VERSION)
GOVERSION := $(shell go version | awk '{print $$3}')
GITCOMMIT := $(shell git log -1 --oneline | awk '{print $$1}')
LDFLAGS = -X main.AppVersion=${APPVERSION} -X main.GoVersion=${GOVERSION} -X main.GitCommit=${GITCOMMIT}
PLATFORM := $(shell uname -s)
ARCH := $(shell uname -m)

.PHONY: clean

build:
	go build -ldflags "${LDFLAGS}" -o bin/restockbot *.go

release:
	go build -ldflags "${LDFLAGS}" -o bin/restockbot-${APPVERSION}-${PLATFORM}-${ARCH} *.go

clean:
	rm -rf bin