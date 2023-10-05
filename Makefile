###Metadata about this makefile and position
MKFILE_PATH := $(lastword $(MAKEFILE_LIST))
CURRENT_DIR := $(patsubst %/,%,$(dir $(realpath $(MKFILE_PATH))))


#Plugin information
PLUGIN_NAME := terraform-provider-venafi-token
PLUGIN_DIR := pkg/bin
DIST_DIR := pkg/dist

# release artifacts must not include the 'v' prefix
ZIP_VERSION := $(shell echo ${VERSION} | cut -c 2-)

TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

# Get the OS dinamically.
# credits to https://gist.github.com/sighingnow/deee806603ec9274fd47
OS_STR :=
CPU_STR :=
ifeq ($(OS),Windows_NT)
	OS_STR := windows
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		CPU_STR := amd64
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		CPU_STR := 386
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		OS_STR := linux

		UNAME_P := $(shell uname -p)
		ifeq ($(UNAME_P),x86_64)
			CPU_STR := amd64
		else
			ifneq ($(filter %86,$(UNAME_P)),)
				CPU_STR := 386
			else
				CPU_STR := amd64
			endif
		endif
	endif
	ifeq ($(UNAME_S),Darwin)
		OS_STR := darwin
		CPU_STR := amd64
	endif
endif

TERRAFORM_TEST_VERSION := 99.9.9
TERRAFORM_TEST_DIR := terraform.d/plugins/registry.terraform.io/venafi/venafi-token/$(TERRAFORM_TEST_VERSION)/$(OS_STR)_$(CPU_STR)

os:
	@echo $(OS_STRING)

all: build test testacc

#Build
build_dev:
	env CGO_ENABLED=0 GOOS=$(OS_STR)   GOARCH=$(CPU_STR) go build -ldflags '-s -w -extldflags "-static"' -a -o $(PLUGIN_DIR)/$(OS_STR)/$(PLUGIN_NAME)_$(VERSION) || exit 1

build:
	env CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags '-s -w -extldflags "-static"' -a -o $(PLUGIN_DIR)/linux/$(PLUGIN_NAME)_$(VERSION) || exit 1
	env CGO_ENABLED=0 GOOS=linux   GOARCH=386   go build -ldflags '-s -w -extldflags "-static"' -a -o $(PLUGIN_DIR)/linux86/$(PLUGIN_NAME)_$(VERSION) || exit 1
	env CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags '-s -w -extldflags "-static"' -a -o $(PLUGIN_DIR)/darwin/$(PLUGIN_NAME)_$(VERSION) || exit 1
	env CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags '-s -w -extldflags "-static"' -a -o $(PLUGIN_DIR)/darwin_arm/$(PLUGIN_NAME)_$(VERSION) || exit 1
	#Build with debugging options, use it for remote debugging. Comment the above line
	#env CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build '-gcflags="all=-N -l" -extldflags "-static"' -a -o $(PLUGIN_DIR)/darwin/$(PLUGIN_NAME)_$(VERSION) || exit 1
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '-s -w -extldflags "-static"' -a -o $(PLUGIN_DIR)/windows/$(PLUGIN_NAME)_$(VERSION).exe || exit 1
	env CGO_ENABLED=0 GOOS=windows GOARCH=386   go build -ldflags '-s -w -extldflags "-static"' -a -o $(PLUGIN_DIR)/windows86/$(PLUGIN_NAME)_$(VERSION).exe || exit 1
	chmod +x $(PLUGIN_DIR)/*

compress:
	$(foreach var,linux linux86 darwin darwin_arm windows windows86,cp LICENSE $(PLUGIN_DIR)/$(var);)
	$(foreach var,linux linux86 darwin darwin_arm windows windows86,cp README.md $(PLUGIN_DIR)/$(var);)
	mkdir -p $(DIST_DIR)
	rm -f $(DIST_DIR)/*
	zip -j "$(DIST_DIR)/${PLUGIN_NAME}_$(ZIP_VERSION)_linux_amd64.zip" $(PLUGIN_DIR)/linux/* || exit 1
	zip -j "$(DIST_DIR)/${PLUGIN_NAME}_$(ZIP_VERSION)_linux_386.zip" $(PLUGIN_DIR)/linux86/* || exit 1
	zip -j "$(DIST_DIR)/${PLUGIN_NAME}_$(ZIP_VERSION)_darwin_amd64.zip" $(PLUGIN_DIR)/darwin/* || exit 1
	zip -j "$(DIST_DIR)/${PLUGIN_NAME}_$(ZIP_VERSION)_darwin_arm64.zip" $(PLUGIN_DIR)/darwin_arm/* || exit 1
	zip -j "$(DIST_DIR)/${PLUGIN_NAME}_$(ZIP_VERSION)_windows_amd64.zip" $(PLUGIN_DIR)/windows/* || exit 1
	zip -j "$(DIST_DIR)/${PLUGIN_NAME}_$(ZIP_VERSION)_windows_386.zip" $(PLUGIN_DIR)/windows86/* || exit 1

collect_artifacts:
	(cd $(DIST_DIR) && sha256sum *.zip > $(PLUGIN_NAME)_$(ZIP_VERSION)_SHA256SUMS)
	(cd $(DIST_DIR) && gpg2 --detach-sign --default-key opensource@venafi.com $(PLUGIN_NAME)_$(ZIP_VERSION)_SHA256SUMS)
	rm -rf artifacts
	mkdir -p artifacts
	cp -rv $(DIST_DIR)/* artifacts

release:
	go get -u github.com/tcnksm/ghr
	ghr -prerelease -n $$RELEASE_VERSION $$RELEASE_VERSION artifacts/

clean:
	rm -fv terraform.tfstate*
	rm -fv $(PLUGIN_NAME)
	rm -rfv $(PLUGIN_DIR)/*
	rm -rfv $(DIST_DIR)/*
	rm -rfv .terraform
	rm -rfv terraform.d
	rm -fv .terraform.lock.hcl

dev: clean fmtcheck
	go test ./...
	env CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags '-s -w -extldflags "-static"' -a -o $(PLUGIN_NAME)_$(VERSION) || exit 1
	terraform init

test: fmtcheck linter test_go testacc test_e2e

test_go:
	go test -v -coverprofile=cov1.out ./venafi
	go tool cover -func=cov1.out

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

# Integration tests using real terraform binary
test_e2e: e2e_init

# This step copies the built terraform plugin to the terraform folder structure, so  changes can be tested.
e2e_init: build_dev
	mkdir -p $(TERRAFORM_TEST_DIR)
	mv $(PLUGIN_DIR)/$(OS_STR)/$(PLUGIN_NAME)_$(VERSION) $(TERRAFORM_TEST_DIR)/$(PLUGIN_NAME)_v$(TERRAFORM_TEST_VERSION)
	chmod 755 $(TERRAFORM_TEST_DIR)/$(PLUGIN_NAME)_v$(TERRAFORM_TEST_VERSION)
	terraform init

linter:
	@golangci-lint --version || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /go/bin
	golangci-lint run
