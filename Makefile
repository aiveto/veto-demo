.DEFAULT_GOAL := demo

ROOT := $(abspath .)
VETO_SRC := $(abspath $(ROOT)/../veto)

export HARBOR_TOKEN := harbor-demo-token
export VETO_APPROVAL_NONCE_DIR := $(ROOT)/.harbor/approvals
export VETO_TOKEN_DIR := $(ROOT)/.harbor/tokens
unexport VETO_APPROVAL_SECRET

.PHONY: demo up show build

demo: build ## Start Harbor, run the walk, stop Harbor
	@mkdir -p "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)" .harbor
	@./bin/harbor > .harbor/harbor.log 2>&1 & echo $$! > .harbor/harbor.pid; \
	./bin/show; \
	status=$$?; \
	kill $$(cat .harbor/harbor.pid) >/dev/null 2>&1 || true; \
	if [ $$status -ne 0 ]; then echo "--- harbor log ---"; cat .harbor/harbor.log; fi; \
	exit $$status

up: build ## Leave the three desks listening and print the MCP command
	@mkdir -p "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)"
	./bin/harbor

show: build ## Run the walk against a Harbor that is already up
	./bin/show

build: ## Build veto, the desks, and the walk
	@test -d "$(VETO_SRC)/cmd/veto" || (echo "need a veto checkout at $(VETO_SRC)"; exit 1)
	@mkdir -p bin "$(VETO_APPROVAL_NONCE_DIR)" "$(VETO_TOKEN_DIR)"
	go build -C "$(VETO_SRC)" -o "$(ROOT)/bin/veto" ./cmd/veto
	go build -o bin/harbor ./cmd/harbor
	go build -o bin/show ./cmd/show
