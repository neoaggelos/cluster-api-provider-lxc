# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Directories.
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
TOOLS_DIR := $(realpath ../../hack/tools)
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
MDBOOK_INSTALL := $(realpath ../../hack/scripts/install-mdbook.sh)
TOC_INSTALL := $(realpath ../../hack/scripts/install-mdbook-toc.sh)

export PATH := $(TOOLS_BIN_DIR):$(PATH)

all: build

# Binaries.
TABULATE := $(TOOLS_BIN_DIR)/mdbook-tabulate
$(TABULATE): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); go build -tags=tools -o $(TOOLS_BIN_DIR)/mdbook-tabulate sigs.k8s.io/cluster-api/hack/tools/mdbook/tabulate

EMBED := $(TOOLS_BIN_DIR)/mdbook-embed
$(EMBED): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); go build -tags=tools -o $(TOOLS_BIN_DIR)/mdbook-embed sigs.k8s.io/cluster-api/hack/tools/mdbook/embed

RELEASELINK := $(TOOLS_BIN_DIR)/mdbook-releaselink
$(RELEASELINK): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); go build -tags=tools -o $(TOOLS_BIN_DIR)/mdbook-releaselink sigs.k8s.io/cluster-api/hack/tools/mdbook/releaselink

GEN_CRD_API_REFERENCE_DOCS := $(TOOLS_BIN_DIR)/gen-crd-api-reference-docs
$(GEN_CRD_API_REFERENCE_DOCS):
	cd $(TOOLS_DIR); go build --tags=tools -o $@ github.com/ahmetb/gen-crd-api-reference-docs

# Only set MDBOOK if it's not set as an environment variable
MDBOOK ?= $(TOOLS_BIN_DIR)/mdbook
$(MDBOOK):
	$(MDBOOK_INSTALL) v0.4.43 $(TOOLS_BIN_DIR)

# Only set TOC if it's not set as an environment variable
TOC ?= $(TOOLS_BIN_DIR)/mdbook-toc
$(TOC):
	$(TOC_INSTALL) 0.14.2 $(TOOLS_BIN_DIR)

.PHONY: generate-api-docs clean-api-docs
generate-api-docs: src/reference/api/v1alpha2/api.md
clean-api-docs: clean-api-docs-v1alpha2

src/reference/api/%/api.md: $(GEN_CRD_API_REFERENCE_DOCS)
	$(GEN_CRD_API_REFERENCE_DOCS) \
		-api-dir=../../api/$* \
		-config=./gen-crd-api-reference-docs/config.json \
		-template-dir=./gen-crd-api-reference-docs/template \
		-out-file=./src/reference/api/$*/api.md

clean-api-docs-%:
	rm src/reference/api/$*/api.md

.PHONY: serve
serve: $(MDBOOK) $(TABULATE) $(EMBED) $(RELEASELINK) $(TOC) generate-api-docs
	$(MDBOOK) serve

.PHONY: build
build: $(MDBOOK) $(TABULATE) $(EMBED) $(RELEASELINK) $(TOC) generate-api-docs
	$(MDBOOK) build
