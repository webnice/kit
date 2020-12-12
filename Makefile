DIR=$(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))

GOPATH := $(GOPATH)
DATE=$(shell date -u +%Y%m%d.%H%M%S.%Z)

default: dep

dep:
	@GO111MODULE=on GOSUMDB=off GOPROXY=direct GOPRIVATE="git.webdesk.ru" go get -u ./...
	@GO111MODULE=on GOSUMDB=off GOPROXY=direct GOPRIVATE="git.webdesk.ru" go mod download
	@GO111MODULE=on GOSUMDB=off GOPROXY=direct GOPRIVATE="git.webdesk.ru" go mod tidy
	@GO111MODULE=on GOSUMDB=off GOPROXY=direct GOPRIVATE="git.webdesk.ru" go mod vendor
.PHONY: dep

clean:
	@echo "cleaning..."
	@rm -rf ${DIR}/src; true
	@rm -rf ${DIR}/bin/*; true
	@rm -rf ${DIR}/pkg/*; true
	@rm -rf ${DIR}/*.log; true
.PHONY: clean
