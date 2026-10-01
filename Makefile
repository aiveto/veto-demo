.DEFAULT_GOAL := demo

ROOT := $(abspath .)
VETO_SRC := $(abspath $(ROOT)/../veto)

export DEMO_TOKEN := demo-token
export VETO_APPROVAL_NONCE_DIR := $(ROOT)/.demo/approvals
export VETO_TOKEN_DIR := $(ROOT)/.demo/tokens
unexport VETO_APPROVAL_SECRET

.PHONY: help demo up show build vet

help: ## Print each target with a one-line description
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-8s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

demo: build ## Start the desks, run the walk, stop the desks
	@mkdir -p "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)" .demo
	@./bin/apis > .demo/apis.log 2>&1 & echo $$! > .demo/apis.pid; \
	./bin/show; \
	status=$$?; \
	kill $$(cat .demo/apis.pid) >/dev/null 2>&1 || true; \
	wait $$(cat .demo/apis.pid) 2>/dev/null || true; \
	if [ $$status -ne 0 ]; then echo "--- desk log ---"; cat .demo/apis.log; fi; \
	exit $$status

up: build ## Leave the three desks listening and print the MCP command
	@mkdir -p "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)"
	./bin/apis

show: build ## Run the walk against desks that are already up
	./bin/show

build: ## Build veto, the desks, and the walk
	@test -d "$(VETO_SRC)/cmd/veto" || (echo "need a veto checkout at $(VETO_SRC)"; exit 1)
	@mkdir -p bin "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)"
	go build -C "$(VETO_SRC)" -o "$(ROOT)/bin/veto" ./cmd/veto
	go build -o bin/apis ./cmd/apis
	go build -o bin/show ./cmd/show

vet: ## Fail on gofmt drift or go vet findings
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then printf '%s\n' "$$out"; exit 1; fi
	go vet ./...
