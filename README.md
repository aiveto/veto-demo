# veto demo

Imagine Harbor sells home goods. Three APIs are already hosted under `app/`: orders, customers, and billing. Veto sits in front of them.

```bash
make demo
```

That needs Go 1.27.1. The `veto` command comes from `github.com/aiveto/veto` in `go.mod`. The module is private, so `make` sets `GOPRIVATE` and fetches it with your GitHub credentials.

Harbor listens on `127.0.0.1:18410` (orders), `:18411` (customers), and `:18412` (billing). Calls want `Authorization: Bearer demo-token`.

## The story

`make demo` plays it.

Harbor has twenty-five operations. The model gets three: search, describe, and invoke.

"Retire order" and "scrap order" both find the delete. Retire is already known. Scrap is a word you added in `app/semantics.yaml`, next to the sentence that this call waits for a person.

Order 10482 is Mara Ellison's wool coat and two cedar trays, $556. Reading it also reads Mara, and her paid invoice, because those links are written in `app/relations.yaml`. The warehouse id on the order is not a link, so nothing calls it.

Jonas Adler's linen throw, order 10490, is cancelled, and that call reaches Harbor. Deleting Mara's order does not. Veto stops and asks a person. Approving it gives a new id. That id deletes the order once, and then the order is gone.

The model never receives the OpenAPI file.

## Your turn

| You want | Open | Run |
| --- | --- | --- |
| Claude, Cursor, or ChatGPT | `app/prompts/claude.md` | `make mcp` |
| The CLI, with no model | | `make cli` |
| A relationship between APIs | `app/relations.yaml` | |
| Semantics, the words for a call | `app/semantics.yaml` | |
| Which contracts veto loads | `app/veto.yaml` | |
| Harbor's OpenAPI | `app/contracts/` | |
| The process hosting Harbor | `app/desks/` | `make mcp` |

`make mcp` leaves Harbor listening and prints the config to paste into Claude, Cursor, or ChatGPT. Paste `app/prompts/claude.md` as the instruction.

That prompt asks the model to retire Mara's order. It reads the order, follows the customer and the invoice, and stops with a pending id. `veto approve` that id prints a second id. Hand the model the second id. The delete runs once. A cancel does not wait. Ask it to cancel order 10490 if you want that from the model. `make demo` already cancelled Jonas.

`make cli` prints the catalog and the joins, then matches "Delete order 10482" and stops before HTTP.

Your own APIs stay where they are. Point `veto.yaml` at your OpenAPI, and depend on `github.com/aiveto/veto` the same way this `go.mod` does.
