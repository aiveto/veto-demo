<img src="docs/veto-social.png" width="1280" alt="veto makes it possible for an AI agent to call your API with context and semantics. A hundred endpoints stay 3 tools.">

Harbor sells home goods. Orders, customers, and billing already run the shop. This repo hosts those three APIs. Veto sits in front. Your APIs stay where they already run.

```bash
make demo
```

Needs Go 1.27.1.

## What you just saw

Twenty-five operations. Three tools: search, describe, invoke. The model never gets the OpenAPI file.

"Retire order" and "scrap order" both find the delete. Scrap is one extra word in `app/semantics.yaml`.

Reading order 10482 also reads Mara Ellison and her paid invoice, because `app/relations.yaml` says so. The warehouse id is on the order and is not a link, so nothing calls it.

Jonas cancels order 10490. That reaches Harbor. Deleting Mara's order does not. Veto holds it. `veto approve` prints a second id. That id deletes once.

## Your turn

```bash
make mcp
```

Harbor stays up. The printed JSON is the MCP config for Claude, Cursor, or ChatGPT. Paste `app/prompts/claude.md` as the instruction. The agent will stop on a pending id. In another shell, with this repo's Makefile env still set:

```bash
veto approve --config app/veto.yaml <pending id>
```

Hand the agent the printed id. The delete runs once.

| You want | Open |
| --- | --- |
| The APIs veto loads | `app/veto.yaml` |
| A link from one API to another | `app/relations.yaml` |
| An extra word for a call | `app/semantics.yaml` |
| Harbor's OpenAPI | `app/contracts/` |
| The CLI, no model | `make cli` |

## On your APIs

Leave them where they are. Write a `veto.yaml` that names your OpenAPI the way `app/veto.yaml` names Harbor's. Add relations when a field should call another operation. Run `veto serve`. The module is [`github.com/aiveto/veto`](https://github.com/aiveto/veto).
