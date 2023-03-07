# Image URL to use all building/pushing image targets
FROM_VERSION ?=v2.3.0
CHART_VERSION ?=2.3.0
CHART_TOVERSION ?=3.0.0
TO_VERSION ?=v3.0.0
TAG ?=v3.0.0
IMGPREFIX ?=radondb/
MYSQL_IMAGE_57 ?=5.7.34
MYSQL_IMAGE_80 ?=8.0.25
MYSQL_IMAGE_57_TAG ?=$(IMGPREFIX)percona-server:$(MYSQL_IMAGE_57)
MYSQL_IMAGE_80_TAG ?=$(IMGPREFIX)percona-server:$(MYSQL_IMAGE_80)
IMG ?= $(IMGPREFIX)mysql-operator:$(TAG)
SIDECAR57_IMG ?= $(IMGPREFIX)mysql57-sidecar:$(TAG)
SIDECAR80_IMG ?= $(IMGPREFIX)mysql80-sidecar:$(TAG)
XENON_IMG ?= $(IMGPREFIX)xenon:$(TAG)
GO_PORXY ?= off
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

update-crd:	## Synchronize the generated YAML files to operator Chart after make manifests.
	make manifests
	cp config/crd/bases/* charts/mysql-operator/crds/

generate: controller-gen generate-go-conversions ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...
	
CONVERSION_GEN := $(shell pwd)/bin/conversion-gen
CODE_GENERATOR_VERSION := $(shell awk '/k8s.io\/client-go/ {print substr($$2, 2)}' go.mod)
conversion-gen: ## Donwload conversion-gen locally if necessary.
	$(call go-get-tool,$(CONVERSION_GEN),k8s.io/code-generator/cmd/conversion-gen@v$(CODE_GENERATOR_VERSION))
generate-go-conversions: conversion-gen $(CONVERSION_GEN) ## Generate conversions go code
	$(CONVERSION_GEN) \
		--input-dirs=./api/v1beta1 \
		--output-file-base=zz_generated.conversion --output-base=. \
		--go-header-file=./hack/boilerplate.go.txt
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: manifests generate fmt vet ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test $(shell go list ./... | grep -v /test/) -coverprofile cover.out

##@ Build

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager ./cmd/manager/main.go
	go build -o bin/sidecar ./cmd/sidecar/main.go
	go build -o bin/nfsbcp ./cmd/nfsbcp/main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/manager/main.go

docker-build:  ## Build docker image with the manager.
	docker buildx build --build-arg GO_PROXY=${GO_PORXY} -t ${IMG} .
	docker buildx build -f Dockerfile.sidecar --build-arg GO_PROXY=${GO_PORXY} -t ${SIDECAR57_IMG} .
	docker buildx build -f build/xenon/Dockerfile --build-arg GO_PROXY=${GO_PORXY} -t ${XENON_IMG} .
	docker buildx build  --build-arg XTRABACKUP_PKG=percona-xtrabackup-80  --build-arg GO_PROXY=${GO_PORXY} -f  Dockerfile.sidecar -t ${SIDECAR80_IMG} .
	docker buildx build  --build-arg XTRABACKUP_PKG=percona-xtrabackup-57  --build-arg GO_PROXY=${GO_PORXY} -f  Dockerfile.sidecar -t ${SIDECAR80_IMG} .
	docker buildx build  --build-arg "MYSQL_IMAGE=${MYSQL_IMAGE_57}"  --build-arg GO_PROXY=${GO_PORXY} -f build/mysql/Dockerfile  -t ${MYSQL_IMAGE_57_TAG}   .
	docker buildx build  --build-arg "MYSQL_IMAGE=${MYSQL_IMAGE_80}"  --build-arg GO_PROXY=${GO_PORXY} -f build/mysql/Dockerfile  -t ${MYSQL_IMAGE_80_TAG}   .
docker-build-mysql57: test ## Build docker image with the manager.
	docker buildx build  --build-arg "MYSQL_IMAGE=${MYSQL_IMAGE_57}"  --build-arg GO_PROXY=${GO_PORXY} -f build/mysql/Dockerfile  -t ${MYSQL_IMAGE_57_TAG}   .
docker-build-mysql80: 
		docker buildx build  --build-arg "MYSQL_IMAGE=${MYSQL_IMAGE_80}"  --build-arg GO_PROXY=${GO_PORXY} -f build/mysql/Dockerfile  -t ${MYSQL_IMAGE_80_TAG}   .
docker-push: ## Push docker image with the manager.
	docker push ${IMG}
	docker push ${SIDECAR_IMG}
	docker push ${XENON_IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -


CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# E2E tests
###########
e2e-local:
	go test -v ./test/e2e  $(G_ARGS) -timeout 20m
todo:
	@grep -Irnw './' -e 'TODO:'|grep -v grep

updateVersion:
	find ./* -type f -name "*.go" -o -name "*.yaml" -exec sed -i "" "s/mysql57-sidecar:$(FROM_VERSION)/mysql57-sidecar:$(TO_VERSION)/g" {} \;
	find ./* -type f -name "*.go" -o -name "*.yaml" -exec sed -i "" "s/xenon:$(FROM_VERSION)/xenon:$(TO_VERSION)/g" {} \;
	find ./* -type f -name "*.go" -o -name "*.yaml" -exec sed -i  "" "s/mysql-operator:$(FROM_VERSION)/mysql-operator:$(TO_VERSION)/g" {} \;
	find ./* -type f -name "*.go" -o -name "*.yaml" -exec sed -i  "" "s/mysql80-sidecar:$(FROM_VERSION)/mysql80-sidecar:$(TO_VERSION)/g" {} \;
	find ./* -type f -name "*.go" -o -name "*.yaml" -exec sed -i  "" "s/mysql-operator-$(FROM_VERSION)/mysql-operator-$(TO_VERSION)/g" {} \;
	find ./* -type f -name "*.go" -o -name "*.yaml" -exec sed -i  "" "s/\"$(FROM_VERSION)\"/\"$(TO_VERSION)\"/g" {} \;
	# sed -i "18s/$(CHART_VERSION)/$(CHART_TOVERSION)/"  charts/mysql-operator/charts/Chart.yaml
	find ./charts/*  -type f -name "*.yaml" -exec sed -i  "" "s/$(CHART_VERSION)/$(CHART_TOVERSION)/g" {} \;
	find ./config/*  -type f -name "*.yaml" -exec sed -i  "" "s/$(CHART_VERSION)/$(CHART_TOVERSION)/g" {} \;
	
CRD_TO_MARKDOWN := $(shell pwd)/bin/crd-to-markdown
CRD_TO_MARKDOWN_VERSION = 0.0.3
crd-to-markdown: ## Download crd-to-markdown locally if necessary.
	$(call go-get-tool,$(CRD_TO_MARKDOWN),github.com/clamoriniere/crd-to-markdown@v$(CRD_TO_MARKDOWN_VERSION))
apidoc: crd-to-markdown $(wildcard api/*/*_types.go)
	$(CRD_TO_MARKDOWN) --links docs/links.csv -f api/v1beta1/mysqlcluster_types.go  -n MySQLCluster > docs/crd_mysqlcluster_v1beta1.md

