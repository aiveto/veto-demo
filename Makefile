.DEFAULT_GOAL := help

ROOT := $(abspath .)

export DEMO_TOKEN := demo-token
export VETO_APPROVAL_NONCE_DIR := $(ROOT)/.demo/approvals
export VETO_TOKEN_DIR := $(ROOT)/.demo/tokens
unexport VETO_APPROVAL_SECRET

.PHONY: help demo cli mcp up build vet

help: ## Print each target with a one-line description
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-8s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

demo: build ## Play the Harbor story over the CLI, then over MCP
	@$(MAKE) --no-print-directory play CMD=./bin/demo-cli
	@$(MAKE) --no-print-directory play CMD=./bin/demo-mcp

cli: build ## Play the story through veto search, describe, and invoke
	@$(MAKE) --no-print-directory play CMD=./bin/demo-cli

mcp: build ## Leave Harbor's APIs listening for Claude, Cursor, or ChatGPT
	@mkdir -p "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)"
	@./bin/desks

up: mcp ## Same as make mcp

play:
	@test -n "$(CMD)" || (echo "CMD required"; exit 1)
	@rm -rf "$(VETO_APPROVAL_NONCE_DIR)"
	@mkdir -p "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)" .demo
	@./bin/desks --quiet > .demo/desks.log 2>&1 & echo $$! > .demo/desks.pid; \
	$(CMD); status=$$?; \
	kill $$(cat .demo/desks.pid) >/dev/null 2>&1 || true; \
	wait $$(cat .demo/desks.pid) 2>/dev/null || true; \
	if [ $$status -ne 0 ]; then echo "--- desk log ---"; cat .demo/desks.log; fi; \
	exit $$status

build: ## Build bin/veto (sibling ../veto if present, else the module in go.mod), then Harbor and both walks
	@mkdir -p bin "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)"
	@if [ -f "$(ROOT)/../veto/go.mod" ]; then \
		go build -C "$(ROOT)/../veto" -o "$(ROOT)/bin/veto" ./cmd/veto; \
	else \
		go build -o bin/veto github.com/aiveto/veto/cmd/veto; \
	fi
	@go build -o bin/desks ./app/cmd/desks
	@go build -o bin/demo-cli ./demo/cli
	@go build -o bin/demo-mcp ./demo/mcp

vet: ## Fail on gofmt drift or go vet findings
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then printf '%s\n' "$$out"; exit 1; fi
	go vet ./...
