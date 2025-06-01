# Image URL to use all building/pushing image targets
TAG ?= latest
REGISTRY ?= ghcr.io
REPO ?= $(REGISTRY)/lxc/cluster-api-provider-incus
IMG ?= $(REPO):$(TAG)

# Used by ko build
export KO_DOCKER_REPO = $(REPO)

# PLATFORMS defines the target platforms for the manager image be built.
PLATFORMS ?= linux/arm64,linux/amd64
LOCAL_PLATFORM ?= linux/$(shell go env GOARCH)

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.31.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

REPO_ROOT := $(shell git rev-parse --show-toplevel)

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	env \
		CLUSTERAPI_CRD_PATHS="$(shell ./hack/dev/capi-crd-paths.sh)" \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
			go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ E2E tests

export CAPI_KUSTOMIZE_PATH ?= $(KUSTOMIZE)
E2E_DATA_DIR ?= $(REPO_ROOT)/test/e2e/data
E2E_CONFIG_PATH ?= $(E2E_DATA_DIR)/config.yaml

E2E_GINKGO_PARALLEL ?= 2
E2E_ARGS ?=
E2E_GINKGO_FOCUS ?=
E2E_GINKGO_SKIP ?=

ifneq ($(strip $(E2E_GINKGO_SKIP)),)
_SKIP_ARGS := $(foreach arg,$(strip $(E2E_GINKGO_SKIP)),-skip="$(arg)")
endif

.PHONY: e2e-image
e2e-image: TAG = e2e ## Build controller image for e2e tests
e2e-image: ko-build

E2E_DATA_DIR ?= $(REPO_ROOT)/test/e2e/data
ARTIFACTS ?= $(REPO_ROOT)/_artifacts

E2E_GINKGO_ARGS ?=
.PHONY: test-e2e
test-e2e: ginkgo kustomize ## Run e2e tests
	time $(GINKGO) \
		-tags=e2e \
		-fail-fast -timeout=3h \
		-v -show-node-events -trace \
		-github-output \
		-nodes=$(E2E_GINKGO_PARALLEL) \
		-focus="$(E2E_GINKGO_FOCUS)" $(_SKIP_ARGS) \
		-output-dir="$(ARTIFACTS)" -junit-report=junit.e2e_suite.1.xml \
		$(E2E_GINKGO_ARGS) \
		./test/e2e/suites/e2e/... -- \
			-config-path="$(E2E_CONFIG_PATH)" \
			-artifacts-folder="$(ARTIFACTS)" \
			-data-folder="$(E2E_DATA_DIR)" \
			$(E2E_ARGS)

.PHONY: run-test-conformance
run-test-conformance: ginkgo kustomize
	env \
		KUBETEST_CONFIGURATION="$(KUBETEST_CONFIGURATION)" \
		time $(GINKGO) \
			-tags=e2e \
			-fail-fast -timeout=3h \
			-v -show-node-events -trace \
			-github-output \
			-focus=conformance \
			$(E2E_GINKGO_ARGS) \
			./test/e2e/suites/conformance/... -- \
				-config-path="$(E2E_CONFIG_PATH)" \
				-artifacts-folder="$(ARTIFACTS)" \
				-data-folder="$(E2E_DATA_DIR)" \
				$(E2E_ARGS)

.PHONY: test-conformance
test-conformance: KUBETEST_CONFIGURATION = $(E2E_DATA_DIR)/kubetest/conformance.yaml
test-conformance: run-test-conformance ## Run conformance tests

.PHONY: test-conformance-fast
test-conformance-fast: KUBETEST_CONFIGURATION = $(E2E_DATA_DIR)/kubetest/conformance-fast.yaml
test-conformance-fast: run-test-conformance ## Run conformance tests (fast)

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
V ?= 0
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go --diagnostics-address=":" --v=${V}

.PHONY: ko-build
ko-build: ko ## Build manager image and load to local docker instance.
	$(KO) build ./cmd --bare --tags $(TAG) --sbom=none --platform=$(LOCAL_PLATFORM) --local

.PHONY: ko-login
ko-login: ko ## Configure credentials for pushing images (needs USERNAME and PASSWORD).
	echo $(PASSWORD) | $(KO) login $(REGISTRY) --username $(USERNAME) --password-stdin

.PHONY: ko-push
ko-push: ko ## Build and push manager image.
	$(KO) build ./cmd --bare --tags $(TAG) --sbom=none --platform=$(PLATFORMS)

##@ Release

.PHONY: release
release: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > dist/infrastructure-components.yaml

.PHONY: dist
dist: release ## Generate release assets.
	cp templates/clusterclass*.yaml dist/
	cp templates/cluster-template*.yaml dist/
	cp metadata.yaml dist/

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GINKGO ?= $(LOCALBIN)/ginkgo
KO ?= $(LOCALBIN)/ko
IMAGE_BUILDER ?= $(LOCALBIN)/image-builder

## Tool Versions
KUSTOMIZE_VERSION ?= v5.5.0
CONTROLLER_TOOLS_VERSION ?= v0.16.4
ENVTEST_VERSION ?= release-0.19
GOLANGCI_LINT_VERSION ?= v1.61.0
KO_VERSION ?= v0.18.0

.PHONY: image-builder
image-builder: $(LOCALBIN)
	go build -o $(IMAGE_BUILDER) ./cmd/exp/image-builder

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.
$(GINKGO): $(LOCALBIN)
	go build -o $@ github.com/onsi/ginkgo/v2/ginkgo

.PHONY: ko
ko: $(KO) ## Download ko locally if necessary.
$(KO): $(LOCALBIN)
	$(call go-install-tool,$(KO),github.com/google/ko,$(KO_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
