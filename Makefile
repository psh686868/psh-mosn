SHELL = /bin/bash

TARGET       = mosnd
CONFIG_FILE  = mosn_config.json
GIT_USER     = alipay
PROJECT_NAME = github.com/${GIT_USER}/psh-mosn

SCRIPT_DIR 	 = $(shell pwd)/etc/script

MAJOR_VERSION = $(shell cat VERSION)
GIT_VERSION   = $(shell git log -1 --pretty=format:%h)
GIT_NOTES     = $(shell git log -1 --oneline)

BUILD_IMAGE   = godep-builder

IMAGE_NAME = ${GIT_USER}/mosnd
REGISTRY = acs-reg.alipay.com

RPM_BUILD_IMAGE = afenp-rpm-builder
RPM_VERSION     = $(shell cat VERSION | tr -d '-')
RPM_TAR_NAME    = afe-${TARGET}
RPM_SRC_DIR     = ${RPM_TAR_NAME}-${RPM_VERSION}
RPM_TAR_FILE    = ${RPM_SRC_DIR}.tar.gz

ut-local:
	go test ./pkg/...

unit-test:
	docker build --rm -t ${BUILD_IMAGE} build/contrib/builder/binary
	docker run --rm -v $(GOPATH):/go -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make ut-local

coverage-local:
	sh ${SCRIPT_DIR}/report.sh

coverage:
	docker build --rm -t ${BUILD_IMAGE} build/contrib/builder/binary
	docker run --rm -v $(GOPATH):/go -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make coverage-local

integrate-local:
	go test ./test/integrate/...

integrate:
	docker build --rm -t ${BUILD_IMAGE} build/contrib/builder/binary
	docker run --rm -v $(GOPATH):/go -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make integrate-local

build:
	docker build --rm -t ${BUILD_IMAGE} build/contrib/builder/binary
	docker run --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make build-local

build-host:
	docker build --rm -t ${BUILD_IMAGE} build/contrib/builder/binary
	docker run --net=host --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make build-local

binary: build

binary-host: build-host

build-local:
	@rm -rf build/bundles/${MAJOR_VERSION}/binary
	CGO_ENABLED=0 go build\
		-ldflags "-B 0x$(shell head -c20 /dev/urandom|od -An -tx1|tr -d ' \n') -X main.Version=${MAJOR_VERSION}(${GIT_VERSION})" \
		-v -o ${TARGET} \
		${PROJECT_NAME}/cmd/mosn/main
	mkdir -p build/bundles/${MAJOR_VERSION}/binary
	mv ${TARGET} build/bundles/${MAJOR_VERSION}/binary
	@cd build/bundles/${MAJOR_VERSION}/binary && $(shell which md5sum) -b ${TARGET} | cut -d' ' -f1  > ${TARGET}.md5
	cp configs/${CONFIG_FILE} build/bundles/${MAJOR_VERSION}/binary

build-linux32:
	@rm -rf build/bundles/${MAJOR_VERSION}/binary
	CGO_ENABLED=0 env GOOS=linux GOARCH=386 go build\
		-ldflags "-B 0x$(shell head -c20 /dev/urandom|od -An -tx1|tr -d ' \n') -X main.Version=${MAJOR_VERSION}(${GIT_VERSION})" \
		-v -o ${TARGET} \
		${PROJECT_NAME}/cmd/mosn/main
	mkdir -p build/bundles/${MAJOR_VERSION}/binary
	mv ${TARGET} build/bundles/${MAJOR_VERSION}/binary
	@cd build/bundles/${MAJOR_VERSION}/binary && $(shell which md5sum) -b ${TARGET} | cut -d' ' -f1  > ${TARGET}.md5
	cp configs/${CONFIG_FILE} build/bundles/${MAJOR_VERSION}/binary

build-linux64:
	@rm -rf build/bundles/${MAJOR_VERSION}/binary
	CGO_ENABLED=0 env GOOS=linux GOARCH=amd64 go build\
		-ldflags "-B 0x$(shell head -c20 /dev/urandom|od -An -tx1|tr -d ' \n') -X main.Version=${MAJOR_VERSION}(${GIT_VERSION})" \
		-v -o ${TARGET} \
		${PROJECT_NAME}/cmd/mosn/main
	mkdir -p build/bundles/${MAJOR_VERSION}/binary
	mv ${TARGET} build/bundles/${MAJOR_VERSION}/binary
	@cd build/bundles/${MAJOR_VERSION}/binary && $(shell which md5sum) -b ${TARGET} | cut -d' ' -f1  > ${TARGET}.md5
	cp configs/${CONFIG_FILE} build/bundles/${MAJOR_VERSION}/binary


image:
	@rm -rf IMAGEBUILD
	cp -r build/contrib/builder/image IMAGEBUILD && cp build/bundles/${MAJOR_VERSION}/binary/${TARGET} IMAGEBUILD && cp -r configs IMAGEBUILD && cp -r etc IMAGEBUILD
	docker build --no-cache --rm -t ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} IMAGEBUILD
	docker tag ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} ${REGISTRY}/${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION}
	rm -rf IMAGEBUILD

rpm:
	@sleep 1  # sometimes device-mapper complains for a relax
	docker build --rm -t ${RPM_BUILD_IMAGE} build/contrib/builder/rpm
	docker run --rm -w /opt/${TARGET}     \
		-v $(shell pwd):/opt/${TARGET}    \
		-e RPM_GIT_VERSION=${GIT_VERSION} \
		-e "RPM_GIT_NOTES=${GIT_NOTES}"   \
	   	${RPM_BUILD_IMAGE} make rpm-build-local

rpm-build-local:
	@rm -rf build/bundles/${MAJOR_VERSION}/rpm
	mkdir -p build/bundles/${MAJOR_VERSION}/rpm/${RPM_SRC_DIR}
	cp -r build/bundles/${MAJOR_VERSION}/binary/${TARGET} build/bundles/${MAJOR_VERSION}/rpm/${RPM_SRC_DIR}
	cp -r build/bundles/${MAJOR_VERSION}/binary/${CONFIG_FILE} build/bundles/${MAJOR_VERSION}/rpm/${RPM_SRC_DIR}
	cp build/contrib/builder/rpm/${TARGET}.spec build/bundles/${MAJOR_VERSION}/rpm
	cp build/contrib/builder/rpm/${TARGET}.service build/bundles/${MAJOR_VERSION}/rpm/${RPM_SRC_DIR}
	cp build/contrib/builder/rpm/${TARGET}.logrotate build/bundles/${MAJOR_VERSION}/rpm/${RPM_SRC_DIR}
	cd build/bundles/${MAJOR_VERSION}/rpm && tar zcvf ${RPM_TAR_FILE} ${RPM_SRC_DIR}
	mv build/bundles/${MAJOR_VERSION}/rpm/${RPM_TAR_FILE} ~/rpmbuild/SOURCES
	chown -R root:root ~/rpmbuild/SOURCES/${RPM_TAR_FILE}
	rpmbuild -bb --clean build/contrib/builder/rpm/${TARGET}.spec             	\
			--define "AFENP_NAME       ${RPM_TAR_NAME}" 					\
			--define "AFENP_VERSION    ${RPM_VERSION}"       		    	\
			--define "AFENP_RELEASE    ${RPM_GIT_VERSION}"     				\
			--define "AFENP_GIT_NOTES '${RPM_GIT_NOTES}'"
	cp ~/rpmbuild/RPMS/x86_64/*.rpm build/bundles/${MAJOR_VERSION}/rpm
	rm -rf build/bundles/${MAJOR_VERSION}/rpm/${RPM_SRC_DIR} build/bundles/${MAJOR_VERSION}/rpm/${TARGET}.spec

shell:
	docker build --rm -t ${BUILD_IMAGE} build/contrib/builder/binary
	docker run --rm -ti -v $(GOPATH):/go -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} /bin/bash

.PHONY: unit-test build image rpm upload shell
