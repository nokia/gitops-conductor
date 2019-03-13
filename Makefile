GO=CGO_ENABLED=0 go
IMAGENAME=gitops-conductor
MYID="$(shell id -u $(whoami)):$(shell id -g $(whoami))"
REGISTRY?=registry.dac.nokia.com/public
VERSION?=latest
PWD=$(shell pwd)

all: ci

test:
	docker run --rm \
    -e DEPCACHEDIR=/tmp/depcache \
    -e "http_proxy=${http_proxy}" \
    -e "https_proxy=${https_proxy}" \
    -e "VERSION=${VERSION}" \
    -e "REGISTRY=${REGISTRY}" \
    -e "IMAGENAME=${IMAGENAME}" \
    -v /var/run/docker.sock:/var/run/docker.sock \
     -v "${PWD}":/go/src/github.com/nokia/gitops-conductor \
                 -w "/go/src/github.com/nokia/gitops-conductor" \
                   operator-build_v0.5.0 /bin/bash -c "go test -v ./pkg/..."

ci: builder operator-build

e2e: bin push

builder:
	docker build -t operator-build_v0.5.0 --build-arg "http_proxy=${http_proxy}" \
	 --build-arg "https_proxy=${https_proxy}" .


bin: baseimg
	operator-sdk build $(REGISTRY)/$(IMAGENAME):${VERSION}
	rm -rf build/_output

baseimg: 
	docker build --build-arg=https_proxy=${http_proxy} --build-arg=http_proxy=${http_proxy} -f build/base/Dockerfile -t gitopsbase build/base

push:
	docker push $(REGISTRY)/$(IMAGENAME):$(VERSION)


operator-build: dep
	docker run --rm \
   -e DEPCACHEDIR=/tmp/depcache \
   -e "http_proxy=${http_proxy}" \
   -e "https_proxy=${https_proxy}" \
   -e "VERSION=${VERSION}" \
   -e "REGISTRY=${REGISTRY}" \
	 -e "IMAGENAME=${IMAGENAME}" \
   -v /var/run/docker.sock:/var/run/docker.sock \
    -v "${PWD}":/go/src/github.com/nokia/gitops-conductor \
                -w "/go/src/github.com/nokia/gitops-conductor" \
                operator-build_v0.5.0 /bin/bash -c "mkdir -p /tmp/depcache/{sources} && cd /go/src/github.com/nokia/gitops-conductor && make bin"

depend:
	dep ensure -v

dep:
	docker run --rm \
                -u ${MYID} \
                -e DEPCACHEDIR=/tmp/depcache \
                -e "http_proxy=${http_proxy}" \
                -e "https_proxy=${https_proxy}" \
                -e "VERSION=${VERSION}" \
                -v /var/run/docker.sock:/var/run/docker.sock \
                -v "${PWD}":/go/src/github.com/nokia/gitops-conductor \
                -w "/go/src/github.com/nokia/gitops-conductor" \
                operator-build_v0.5.0  /bin/bash -c "mkdir -p /tmp/depcache/{sources} && pwd &&  ls -l && dep ensure -v"

k8s:
	operator-sdk generate k8s
