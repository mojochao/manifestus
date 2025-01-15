# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL := /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# -----------------------------------------------------------------------------
# Define build variables
# -----------------------------------------------------------------------------

# Set app identity.
APP_NAME ?= manifestus
APP_VERSION ?= $(shell cat VERSION)

# Set Docker image build configuration.
DOCKERFILE ?= Dockerfile

# Set documentation preview configuration.
DOCS_DIR = docs
DOCS_PORT ?= 8000

# Set git repo identity
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT ?= $(shell git rev-parse HEAD)
GIT_ABBREV ?= $(shell git rev-parse --short HEAD)
GIT_DIRTY  ?= $(shell test -n "`git status --porcelain`" && echo "-dirty" || true)
GIT_TAG = "$(GIT_ABBREV)$(GIT_DIRTY)"

# Set go compiler and linker flags.
GO_PACKAGE ?= github.com/mojochao/$(APP_NAME)
GO_LDFLAGS = -ldflags "-X $(GO_PACKAGE)/core.Version=$(APP_VERSION)"

# Set image registry configuration.
REGISTRY_HOSTNAME = ghcr.io
REGISTRY_USERNAME = mojochao
REGISTRY_REPONAME = $(REGISTRY_HOSTNAME)/$(REGISTRY_USERNAME)/$(APP_NAME)

# Set image identity configuration.
IMAGE_ABBREV  = $(REGISTRY_REPONAME):$(GIT_TAG)      # tag all images with short git commit hash and '-dirty' suffix if in dirty state
IMAGE_COMMIT  = $(REGISTRY_REPONAME):$(GIT_COMMIT)   # tag all images with full git commit hash only if not in dirty state
IMAGE_VERSION = $(REGISTRY_REPONAME):v$(APP_VERSION) # tag release images with app version as well
IMAGE_LATEST  = $(REGISTRY_REPONAME):latest 		   # tag latest release images with 'latest' as well

# -----------------------------------------------------------------------------
# Define build targets
# -----------------------------------------------------------------------------

# Display help information by default.
.DEFAULT_GOAL := help

##@ Info targets

# The 'help' target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
#
# See https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters for more
# info on the usage of ANSI control characters for terminal formatting.
#
# See http://linuxcommand.org/lc3_adv_awk.php for more info on the awk command.

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: vars
vars: ## Show environment variables used by this Makefile
	@echo "APP_NAME:          $(APP_NAME)"
	@echo "APP_VERSION:       $(APP_VERSION)"
	@echo "GIT_COMMIT:        $(GIT_COMMIT)"
	@echo "GIT_TAG:           $(GIT_TAG)"
	@echo "GO_PACKAGE:        $(GO_PACKAGE)"
	@echo "GO_LDFLAGS:        $(GO_LDFLAGS)"
	@echo "REGISTRY_HOSTNAME: $(REGISTRY_HOSTNAME)"
	@echo "REGISTRY_USERNAME: $(REGISTRY_USERNAME)"
	@echo "REGISTRY_REPONAME: $(REGISTRY_REPONAME)"
	@echo "IMAGE_ABBREV:      $(IMAGE_ABBREV)"
ifeq ($(GIT_DIRTY),)
	@echo "IMAGE_COMMIT:      $(IMAGE_COMMIT)"
endif
ifneq ($(RELEASE),)
	@echo "IMAGE_VERSION:     $(IMAGE_VERSION)"
	@echo "IMAGE_LATEST:      $(IMAGE_LATEST)"
endif

##@ Application targets

.PHONY: build
build: ## Build application binary
	@echo
	@echo 'building $(APP_NAME) ...'
	@CGO_ENABLED=0 go build $(GO_LDFLAGS) -tags netgo,osusergo -o ./bin/$(APP_NAME) .

.PHONY: lint
lint: ## Lint application source
	@echo
	@echo 'linting $(APP_NAME) ...'
	@golangci-lint run

.PHONY: test
test: ## Run application tests
	@echo
	@echo 'testing $(APP_NAME) ...'
	@go test -v ./...

.PHONY: clean
clean: ## Clean application binary
	@echo
	@echo 'cleaning $(APP_NAME) build artifacts ...'
	@rm -rf ./bin/${APP_NAME}

##@ Documentation targets

.PHONY: docs
docs: ## Preview application documentation
	@echo
	@echo 'previewing $(APP_NAME) documentation at http://localhost:$(DOCS_PORT)/ ...'
	@python3 -m http.server $(DOCS_PORT) -d $(DOCS_DIR)

##@ Image targets

.PHONY: image
image: ## Build container image
	@echo
	@echo 'building image $(IMAGE_ABBREV) ...'
	DOCKER_BUILDKIT=1 DOCKER_CLI_HINTS=false docker build -t $(IMAGE_ABBREV) .
ifeq ($(GIT_DIRTY),)
	@echo 'tagging clean image $(IMAGE_COMMIT) ...'
	docker tag $(IMAGE_ABBREV) $(IMAGE_COMMIT)
endif
ifneq ($(RELEASE),)
	@echo
	@echo 'tagging version image v$(IMAGE_VERSION)'
	docker tag $(IMAGE_COMMIT) v$(IMAGE_VERSION)
	@echo
	@echo 'tagging latest image $(IMAGE_LATEST)'
	docker tag $(IMAGE_VERSION) $(IMAGE_LATEST)
endif
