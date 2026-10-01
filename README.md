# veto demo

Imagine Harbor sells home goods. Three APIs are already hosted under `app/`: orders, customers, and billing. Veto sits in front of them.

```bash
make demo
```

That needs Go 1.27.1. The `veto` command comes from `github.com/aiveto/veto` in `go.mod`. The module is private, so `make` sets `GOPRIVATE` and fetches it with your GitHub credentials.

Harbor listens on `127.0.0.1:18410` (orders), `:18411` (customers), and `:18412` (billing). Calls want `Authorization: Bearer demo-token`.

## What `make demo` prints

The source of this printout is `walk/`. There is no separate command.

```text
veto
Harbor's orders, customers, and billing are up. Veto is in front of them.
The model proposes. Veto decides.

1  Catalog
    ok: catalog (25 operations)
    orders.get --[Order.customerId]--> customers.get
    orders.get --[Order.invoiceId]--> invoices.get
    Doctor loaded Harbor's APIs. orders.delete is waiting on a person.

2  Three tools
    capabilities_search, capabilities_describe, capabilities_invoke
    Harbor publishes more than twenty operations. The agent sees these three.

3  Semantics
    "retire order"  →  orders.delete
    "scrap order"  →  orders.delete
    retire is the built-in synonym. scrap is the overlay. The sentence says the call waits.

4  Context pack
    The pack for "who placed order 10482" has no raw spec in it.

5  Relations
    orders.get  →  customers.get   by Order.customerId
    orders.get  →  invoices.get    by Order.invoiceId
    10482  Mara Ellison  wool coat and two cedar trays  $556.00  invoice paid
    warehouse wh_sfo_1 was on the order. Nothing called it.

6  Cancel reaches the desk
    orders.cancel 10490 went to the desk. No approval. Jonas Adler's linen throw is cancelled.

7  Replay stops
    "Delete order 10482" selected orders.delete.
    The trace says confirmation_required. The desk saw no DELETE.

8  MCP stops the same way
    capabilities_invoke orders.delete  →  confirmation_required
    pending <pending-id>

9  A person says yes
    veto approve  →  <second-id>
    One DELETE /orders/10482. The same id does not run again. The order is gone.

10 Eval
    veto eval  delete still requires confirmation, and the pack still omits the spec.

Mara's order was read, then her customer, then her invoice.
The warehouse id stayed a field.
Cancel reached the desk. Delete did not, until a person approved it.
The approved id ran once.
```

The two ids change every run. The pending id is the handle. The second id is the one that deletes the order, once.

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
