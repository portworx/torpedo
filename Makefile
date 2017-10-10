TORPEDO_IMG=$(DOCKER_HUB_REPO)/$(DOCKER_HUB_TORPEDO_IMAGE):$(DOCKER_HUB_TAG)

ifndef TAGS
TAGS := daemon
endif

ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'github.com/portworx/torpedo/vendor' | grep -v 'github.com/portworx/torpedo/tests')
endif

ifeq ($(BUILD_TYPE),debug)
BUILDFLAGS := -gcflags "-N -l"
endif

BASE_DIR := $(shell git rev-parse --show-toplevel)

BIN :=$(BASE_DIR)/bin
GOFMT := gofmt

.DEFAULT_GOAL=all

all: vet lint build fmt

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v $(PKGS)

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f $(PKGS)

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t $(PKGS)

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -tags "$(TAGS)" -d -v -t -u -f $(PKGS)

fmt:
	@echo -e "Performing gofmt on following: $(PKGS)"
	@./scripts/do-gofmt.sh $(PKGS)

build:
	mkdir -p $(BIN)
	go build -tags "$(TAGS)" $(BUILDFLAGS) $(PKGS)

	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega
	ginkgo build -r

	find . -name '*.test' | awk '{cmd="cp  "$$1"  $(BIN)"; system(cmd)}'
	chmod -R 755 bin/*

install:
	go install -tags "$(TAGS)" $(PKGS)

lint:
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v vendor | grep -v '\.pb\.go' | grep -v '\.pb\.gw\.go'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet:
	go vet $(PKGS)

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck -tags "$(TAGS)" $(PKGS)

pretest: lint vet errcheck

test:
	go test -tags "$(TAGS)" $(TESTFLAGS) $(PKGS)

container:
	@echo "Building container: docker build --tag $(TORPEDO_IMG) -f Dockerfile ."
	sudo docker build --tag $(TORPEDO_IMG) -f Dockerfile .

deploy: container
	sudo docker push $(TORPEDO_IMG)

docker-build:
	docker build -t torpedo/docker-build -f Dockerfile.build .
	@echo "Building torpedo using docker"
	docker run \
		--privileged \
		-v $(shell pwd):/go/src/github.com/portworx/torpedo \
		-e DOCKER_HUB_REPO=$(DOCKER_HUB_REPO) \
		-e DOCKER_HUB_TORPEDO_IMAGE=$(DOCKER_HUB_TORPEDO_IMAGE) \
		-e DOCKER_HUB_TAG=$(DOCKER_HUB_TAG) \
		torpedo/docker-build make all

clean:
	-sudo rm -rf bin
	@echo "Deleting image "$(TORPEDO_IMG)
	-docker rmi -f $(TORPEDO_IMG)
	go clean -i $(PKGS)
