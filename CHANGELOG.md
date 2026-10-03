# Changelog

## Unreleased

- README and the Claude prompt say `veto approve` shares `VETO_APPROVAL_NONCE_DIR` with `make mcp`. A sibling `../veto` checkout is what `make` builds. CI runs `make demo`.

## v0.1.0 - 2026-10-02

- Tracks veto v0.1.0. `make demo` does not set `GOPRIVATE`.
- Harbor's orders, customers, and billing run under `app/`. `make demo` plays the walk. `make mcp` leaves them listening.
