NAME=arm
BINARY=packer-plugin-${NAME}

HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)


build:
	go build -o ${BINARY}

install-packer-sdc: ## Install packer software development command
	go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@${HASHICORP_PACKER_PLUGIN_SDK_VERSION}

ci-release-docs: install-packer-sdc
	packer-sdc renderdocs -src docs -partials docs-partials/ -dst docs/
	/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"

plugin-check: install-packer-sdc
	@PATH="$${PATH}:~/go/bin" \
	packer-sdc plugin-check ${BINARY}

lint:
	golangci-lint run
	pre-commit run -a

.PHONY: all $(MAKECMDGOALS)
