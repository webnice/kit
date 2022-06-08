DIR=$(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))

OLDGOPATH   := $(GOPATH)
GOPATH      := $(GOPATH)
DATE         = $(shell date -u +%Y%m%d.%H%M%S.%Z)
GOGENERATE   = $(shell if [ -f .gogenerate ]; then cat .gogenerate; fi)

default: dep

dep:
	@go get -u ...
.PHONY: dep

gen:
	@for PKGNAME in $(GOGENERATE); do go generate $${PKGNAME}; done
.PHONY: dep

clean:
	@echo "cleaning..."
	@rm -rf ${DIR}/src; true
	@rm -rf ${DIR}/bin/*; true
	@rm -rf ${DIR}/pkg/*; true
	@rm -rf ${DIR}/*.log; true
.PHONY: clean
