# SPDX-License-Identifier: Apache-2.0
#
# Copyright (c) 2024  Panasonic Automotive Systems, Co., Ltd.
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
#

#CURDIR := $(dir $(lastword $(MAKEFILE_LIST)))

GO?=go
GOBUILDFLAGS?=-v

THIS_DIR=.

MODULES=

INSTALL_MODULES=$(patsubst %,install-%, $(MODULES))
CLEAN_MODULES=$(patsubst %,clean-%, $(MODULES))
TEST_MODULES=$(patsubst %,test-%, $(MODULES))
FMT_MODULES=$(patsubst %,fmt-%, $(MODULES))
LINT_MODULES=$(patsubst %,lint-%, $(MODULES))
DOC_MODULES=$(patsubst %,doc-%, $(MODULES))

ULANODESO?=libulanode.so
ULANODEH?=libulanode.h
INSTALL_LIBDIR?=$(GOPATH)/pkg/libulanode


.PHONY: linuxulanodelib        
linuxulanodelib:
	set -e ; \
        GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 CC=${CC} $(GO) build -buildmode=c-shared -o $(ULANODESO) $(GOBUILDFLAGS) -ldflags '-extldflags "-Wl,-soname=$(ULANODESO)"' . ; \
        install -o $(USER) -g $(USER) -m 755 -d $(INSTALL_LIBDIR) ; \
        install -o $(USER) -g $(USER) -m 664 $(ULANODESO) $(INSTALL_LIBDIR)/$(ULANODESO) ; \
        install -o $(USER) -g $(USER) -m 664 $(ULANODEH) $(INSTALL_LIBDIR)/$(ULANODEH) ; \
        rm $(ULANODESO) $(ULANODEH)


P.HONY: all install
all: install

.PHONY: $(INSTALL_MODULES)
$(INSTALL_MODULES):
	set -e;\
	target=`echo $@ | sed -e 's/install-//'`;\
	make -C $${target} install

.PHONY: $(TEST_MODULES)
$(TEST_MODULES) :
	set -e;\
	target=`echo $@ | sed -e 's/test-//'`;\
	make -C $${target} test

.PHONY: $(FMT_MODULES)
$(FMT_MODULES) :
	set -e;\
	target=`echo $@ | sed -e 's/fmt-//'`;\
	make -C $${target} fmt

.PHONY: $(LINT_MODULES)
$(LINT_MODULES):
	set -e;\
	target=`echo $@ | sed -e 's/lint-//'`;\
	make -C $${target} lint

.PHONY: $(DOC_MODULES)
$(DOC_MODULES):
	set -e;\
	target=`echo $@ | sed -e 's/doc-//'`;\
	make -C $${target} doc

.PHONY: $(CLEAN_MODULES)
$(CLEAN_MODULES):
	set -e;\
	target=`echo $@ | sed -e 's/clean-//'`;\
	make -C $${target} clean 

install: $(INSTALL_MODULES)
	set -e;\
	$(GO) install ${GOBUILDFLAGS} .

.PHONY: test
test: $(TEST_MODULES)
	set -e;\
	$(GO) test .

.PHONY: fmt
fmt: $(FMT_MODULES)
	set -e;\
	$(GO) fmt .

.PHONY: lint
lint: $(LINT_MODULES)
	set -e;\
	$(GO) vet .

.PHONY: doc
doc: $(DOC_MODULES)
	set -e;\
	$(GO) doc .

.PHONY: clean
clean: $(CLEAN_MODULES)
	set -e;\
	$(GO) clean -v .


