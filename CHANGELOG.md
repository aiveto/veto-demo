# Changelog

## Unreleased

- `make` and `make help` print the targets. `make demo` still plays the story.

## v0.1.5 - 2026-10-03

- Tracks veto v0.1.5.

## v0.1.4 - 2026-10-03

- Tracks veto v0.1.4.

## v0.1.3 - 2026-10-03

- Tracks veto v0.1.3. Search returns the pack line, not the operation.

## v0.1.2 - 2026-10-03

- Tracks veto v0.1.2.
- `GET /orders` honors `status` and `limit`.

## v0.1.1 - 2026-10-03

- Tracks veto v0.1.1. `veto approve --config` is on that binary.

- README and the Claude prompt say `veto approve` shares `VETO_APPROVAL_NONCE_DIR` with `make mcp`. A sibling `../veto` checkout is what `make` builds. CI runs `make demo`.

## v0.1.0 - 2026-10-02

- Tracks veto v0.1.0. `make demo` does not set `GOPRIVATE`.
- Harbor's orders, customers, and billing run under `app/`. `make demo` plays the walk. `make mcp` leaves them listening.
