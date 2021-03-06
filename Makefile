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
	$(GOENV) go build -ldflags '-w' -o wb-mqtt-astra cmd/wb-mqtt-astra/main.go

install:
	mkdir -p $(DESTDIR)/usr/bin/ $(DESTDIR)/etc/init.d/ $(DESTDIR)/etc/default/
	install -m 0644 initscripts/defaults $(DESTDIR)/etc/default/wb-mqtt-astra
	install -m 0755 wb-mqtt-astra $(DESTDIR)/usr/bin/
	install -m 0755 initscripts/wb-mqtt-astra $(DESTDIR)/etc/init.d/wb-mqtt-astra

AUTHOR_NAME = $(shell git config user.name)
AUTHOR_EMAIL = $(shell git config user.email)
DATE = $(shell date '+%a, %d %b %Y %T %z')
RELEASE_URGENCY = low
RELEASE_DEBIAN_TARGET = wheezy
release:
	@git fetch --tags
	@echo 'Changes:' && git log --format="* %s" `git describe --tags --abbrev=0`..HEAD | cat; echo ''
	@echo 'What version is it?' && read version && \
	  echo "wb-mqtt-astra ($$version) $(RELEASE_DEBIAN_TARGET); urgency=$(RELEASE_URGENCY)" > debian/changelog_tmp &&\
	  echo >> debian/changelog_tmp &&\
	  git log --format="  * %s" `git describe --tags --abbrev=0`..HEAD >> debian/changelog_tmp &&\
	  echo >> debian/changelog_tmp &&\
	  echo ' -- $(AUTHOR_NAME) <$(AUTHOR_EMAIL)>  $(DATE)' >> debian/changelog_tmp &&\
	  echo >> debian/changelog_tmp && cat debian/changelog >> debian/changelog_tmp &&\
	  mv debian/changelog_tmp debian/changelog &&\
	  git add debian/changelog && git commit -m "add changelog for release v$${version}" && git push &&\
	  SSH_AUTH_SOCK= WBDEV_TARGET=$(RELEASE_DEBIAN_TARGET)-armel wbdev gdeb &&\
	  $(MAKE) build_amd64_linux &&\
	  echo dh_auto_configure >debian/wb-mqtt-astra.debhelper.log &&\
	  echo dh_auto_build >debian/wb-mqtt-astra.debhelper.log &&\
	  docker run --rm -v `pwd`/..:/deb virus/checkinstall:wheezy bash -c \
	    "cd /deb/wb-mqtt-astra && fakeroot debian/rules binary" && rm -rf _build && \
	  package_cloud push wb-mqtt-astra/main/debian/$(RELEASE_DEBIAN_TARGET) \
	    ../wb-mqtt-astra_$${version}_armel.deb ../wb-mqtt-astra_$${version}_amd64.deb &&\
	  hub release create -a ../wb-mqtt-astra_$${version}_armel.deb \
	    -a ../wb-mqtt-astra_$${version}_amd64.deb v$$version &&\
	  rm ../wb-mqtt-astra_$${version}*

build_amd64_linux:
	mkdir -p _build/go/src/github.com/andrey-yantsen/ && \
	  ln -s `pwd` _build/go/src/github.com/andrey-yantsen/wb-mqtt-astra && \
	  GOPATH=`pwd`/_build/go && go get github.com/contactless/wbgo && \
	  cd _build/go/src/github.com/andrey-yantsen/wb-mqtt-astra && \
	  $(MAKE) DEB_TARGET_ARCH=amd64 GOARCH=amd64 GOOS=linux deps build
