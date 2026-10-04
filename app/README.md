# Harbor's APIs

The three APIs that run Harbor's online business: orders, customers, and billing. Veto is `bin/veto` one directory up, built from the module in `go.mod`.

| You want | File |
| --- | --- |
| A link from one API to another | `relations.yaml` |
| The words for a call | `semantics.yaml` |
| The file that names the APIs | `veto.yaml` |
| The OpenAPI | `contracts/` |
| This process | `desks/` |
| The instruction for Claude, Cursor, or ChatGPT | `prompts/claude.md` |
| The instruction for a skill on the CLI | `prompts/skill.md` |
| The checks `veto eval` runs | `cases/` |

From the repo root, `make cli` plays the story through search, describe, and invoke. `make mcp` leaves these APIs listening and prints the config to paste. `make demo` runs both. `veto approve` must use the same `VETO_APPROVAL_NONCE_DIR` the printed MCP env sets.
