# Changelog

## Unreleased

## v0.1.12 - 2026-10-08

- Tracks veto v0.1.12.

## v0.1.11 - 2026-10-08

- Tracks veto v0.1.11.

## v0.1.10 - 2026-10-07

- Tracks veto v0.1.10.

## v0.1.9 - 2026-10-07

- Tracks veto v0.1.9.

## v0.1.8 - 2026-10-06

- Tracks veto v0.1.8. The Harbor walk calls the customer and the invoice from `next_calls` instead of hardcoded ids.

## v0.1.7 - 2026-10-04

- Tracks veto v0.1.7. Query, path, and header values are checked before HTTP. Preview keeps integer digits. CLI, JSON, and MCP are locked by one surface test.

## v0.1.6 - 2026-10-04

- Tracks veto v0.1.6. Search hits include `related_calls`. A pending id submitted as `approval_id` is `pending_approval`.
- `make cli` plays the story through search, describe, and invoke. `make mcp` leaves the APIs up for a host. `make demo` runs both. `app/prompts/skill.md` is the CLI instruction.

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
