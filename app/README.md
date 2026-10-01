# Harbor's APIs

The three APIs that run Harbor's online business: orders, customers, and billing. Veto is `bin/veto` one directory up, built from the module in `go.mod`.

| You want | File |
| --- | --- |
| A relationship between APIs | `relations.yaml` |
| Semantics, the words for a call | `semantics.yaml` |
| Which contracts veto loads | `veto.yaml` |
| The OpenAPI | `contracts/` |
| This process | `desks/` |
| The instruction for Claude, Cursor, or ChatGPT | `prompts/claude.md` |
| The checks `veto eval` runs | `cases/` |

From the repo root, `make demo` runs it. `make mcp` leaves these APIs listening and prints the config to paste. `make cli` runs the veto CLI.
