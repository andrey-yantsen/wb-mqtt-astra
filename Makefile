GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOENV_EXTRA :=

DEB_TARGET_ARCH ?=

ifeq ($(DEB_TARGET_ARCH),armel)
GOARCH := arm
GOOS := linux
GOENV_EXTRA := GOARM=5 CC_FOR_TARGET=arm-linux-gnueabi-gcc CC=$$CC_FOR_TARGET CGO_ENABLED=1
endif
ifeq ($(DEB_TARGET_ARCH),armhf)
GOARCH := arm
GOOS := linux
GOENV_EXTRA := GOARM=6 CC_FOR_TARGET=arm-linux-gnueabihf-gcc CC=$$CC_FOR_TARGET CGO_ENABLED=1
endif

CURDIR := $(shell pwd)
GOENV := GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOENV_EXTRA)
GOVER := $(shell go version | cut -d ' ' -f 3)

all: build

deps:
	cd $$($(GOENV) go env GOPATH)\
	  && wget -O $(GOOS)_$(GOARCH).tar.gz $$(curl -s https://api.github.com/repos/andrey-yantsen/teko-astra-go/releases/latest | fgrep browser_download_url | cut -d'"' -f 4 | fgrep $(GOVER)_$(GOOS)_$(GOARCH).)\
	  && tar -zxf $(GOOS)_$(GOARCH).tar.gz\
	  && rm $(GOOS)_$(GOARCH).tar.gz

build:
	$(GOENV) go build -o wb-mqtt-astra

install:
	mkdir -p $(DESTDIR)/usr/bin/ $(DESTDIR)/etc/init.d/ $(DESTDIR)/etc/default/
	install -m 0644 initscripts/defaults $(DESTDIR)/etc/default/wb-mqtt-astra
	install -m 0755 wb-mqtt-astra $(DESTDIR)/usr/bin/
	install -m 0755 initscripts/wb-mqtt-astra $(DESTDIR)/etc/init.d/wb-mqtt-astra