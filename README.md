# veto demo

Imagine Harbor sells home goods. Three APIs are already hosted under `app/`: orders, customers, and billing. Veto sits in front of them.

```bash
make demo
```

That needs Go 1.27.1. The `veto` command comes from `github.com/aiveto/veto` in `go.mod`. The module is private, so `make` sets `GOPRIVATE` and fetches it with your GitHub credentials.

You watch. The walk reads Mara's order, cancels someone else's, and approves Mara's delete. Harbor is on `127.0.0.1:18410` (orders), `:18411` (customers), and `:18412` (billing). Calls want `Authorization: Bearer demo-token`.

## What the walk prints

**Catalog.** Twenty-five operations. `orders.delete` is waiting on a person.

**Three tools.** The agent sees `capabilities_search`, `capabilities_describe`, and `capabilities_invoke`.

**Semantics.** "retire order" and "scrap order" both name `orders.delete`. Retire is built in. Scrap is one line in `app/semantics.yaml`. The sentence says the call waits for a person.

**Context pack.** "Who placed order 10482" does not put the OpenAPI file in front of the model.

**Relations.** Order 10482 is Mara Ellison: a wool coat and two cedar trays, $556.00. The customer is Mara. The invoice is paid. `warehouseId` `wh_sfo_1` stays on the order. The joins are in `app/relations.yaml`: `customerId` calls `customers.get`, `invoiceId` calls `invoices.get`.

**Cancel.** Jonas Adler's order 10490, a linen throw, is cancelled. That call reaches Harbor. A cancel does not wait for a person.

**Delete.** "Delete order 10482" selects `orders.delete` and stops. Harbor sees no DELETE. The MCP call stops the same way and returns a pending id. `veto approve` prints a second id. That id deletes the order once. The order is gone.

**Eval.** The same stop is a check you can run again.

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

That prompt asks the model to retire Mara's order. It reads the order, follows the customer and the invoice, and stops with a pending id. `veto approve` that id prints a second id. Hand the model the second id. The delete runs once. A cancel does not wait. Ask it to cancel order 10490 if you want that from the model. The walk already cancelled Jonas.

`make cli` prints the catalog and the joins, then matches "Delete order 10482" and stops before HTTP.

`walk/` is the program `make demo` runs. Your own APIs stay where they are. Point `veto.yaml` at your OpenAPI, and depend on `github.com/aiveto/veto` the same way this `go.mod` does.
