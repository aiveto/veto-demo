# Harbor's APIs

Imagine these three are already hosted. Veto is the `bin/veto` binary one directory up, built from the module in `go.mod`. It sits in front of them.

| You want | File |
| --- | --- |
| Add a relationship between APIs | `relations.yaml` |
| Add semantics, the words for a call | `semantics.yaml` |
| Point veto at other contracts | `veto.yaml` |
| Read the OpenAPI | `contracts/` |
| See the process that hosts them here | `desks/` |
| Tell Claude, Cursor, or ChatGPT what to do | `prompts/claude.md` |
| The checks the CLI runs | `cases/` |

From the repo root: `make demo` plays the story, `make cli` runs the veto CLI and the scripted agent, `make mcp` leaves these APIs listening.
