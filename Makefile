MAJOR  = 0
MINOR  = 8
BUGFIX = 1

EXE_NAME = fastlistfiles

VERSION = $(MAJOR).$(MINOR).$(BUGFIX)

BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_HOST = $(shell hostname)
GIT_HASH = $(shell git rev-parse HEAD)
GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
GIT_USER = $(shell git config --global user.name)
GIT_EMAIL = $(shell git config --global user.email)

LD_FLAGS = -X main.version=${VERSION} -X main.scmHash=${GIT_HASH} -X main.scmBranch=${GIT_BRANCH} -X main.buildDate=${BUILD_DATE} -X main.buildHost=${BUILD_HOST} -X main.scmEmail=${GIT_EMAIL}

default:
	go build -v -a  -ldflags "$(LD_FLAGS)"

release: clean
	GOOS=darwin go build -ldflags "-s -w $(LD_FLAGS)" -a -v -o $(EXE_NAME)-darwin
	GOOS=linux go build -ldflags "-s -w $(LD_FLAGS)" -a -v -o $(EXE_NAME)-linux

linux:
	GOOS=linux go build -ldflags "$(LD_FLAGS)" -a -v -o $(EXE_NAME)-linux

linux-zip: linux
	zip $(EXE_NAME)-linux-v$(VERSION).zip $(EXE_NAME)-linux
	zipcloak $(EXE_NAME)-linux-v$(VERSION).zip

clean:
	rm -f $(EXE_NAME)
	rm -f $(EXE_NAME)*.zip
	rm -f $(EXE_NAME)-darwin*
	rm -f $(EXE_NAME)-linux*
