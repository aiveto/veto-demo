.DEFAULT_GOAL := demo

ROOT := $(abspath .)

export GOPRIVATE := github.com/aiveto/veto
export DEMO_TOKEN := demo-token
export VETO_APPROVAL_NONCE_DIR := $(ROOT)/.demo/approvals
export VETO_TOKEN_DIR := $(ROOT)/.demo/tokens
unexport VETO_APPROVAL_SECRET

.PHONY: help demo cli mcp up show build vet

help: ## Print each target with a one-line description
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-8s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

demo: build ## Play the whole story against Harbor's APIs
	@mkdir -p "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)" .demo
	@./bin/desks > .demo/desks.log 2>&1 & echo $$! > .demo/desks.pid; \
	./bin/show; \
	status=$$?; \
	kill $$(cat .demo/desks.pid) >/dev/null 2>&1 || true; \
	wait $$(cat .demo/desks.pid) 2>/dev/null || true; \
	if [ $$status -ne 0 ]; then echo "--- desk log ---"; cat .demo/desks.log; fi; \
	exit $$status

cli: build ## Run the veto CLI on Harbor's contracts
	./bin/veto validate --config app/veto.yaml
	./bin/veto replay --config app/veto.yaml --message "Delete order 10482"

mcp: build ## Leave Harbor's APIs listening for Claude, Cursor, or ChatGPT
	@mkdir -p "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)"
	./bin/desks

up: mcp ## Same as make mcp

show: build ## Play the story again while Harbor's APIs are already up
	./bin/show

build: ## Build bin/veto from the module in go.mod, then Harbor's APIs and the walk
	@mkdir -p bin "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)"
	go build -o bin/veto github.com/aiveto/veto/cmd/veto
	go build -o bin/desks ./app/cmd/desks
	go build -o bin/show ./walk

vet: ## Fail on gofmt drift or go vet findings
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then printf '%s\n' "$$out"; exit 1; fi
	go vet ./...
