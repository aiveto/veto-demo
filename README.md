# veto demo

Harbor sells home goods. Three APIs are already hosted under `app/`: orders, customers, and billing. Veto sits in front of them. This is what veto does with APIs you already have.

```bash
make demo
```

That needs a checkout of [veto](https://github.com/aiveto/veto) next to this directory, and Go 1.27.1. `make` builds `bin/veto` from that checkout.

## Where to go

| You want | Open | Run |
| --- | --- | --- |
| The whole story | | `make demo` |
| The CLI | | `make cli` |
| Claude, Cursor, or ChatGPT | `app/prompts/claude.md` | `make mcp` |
| The agent, with no live model | | `make cli` |
| A relationship between APIs | `app/relations.yaml` | `make demo` |
| Semantics, the words for a call | `app/semantics.yaml` | `make demo` |
| Which contracts veto loads | `app/veto.yaml` | |
| Harbor's API contracts | `app/contracts/` | |
| The process hosting those APIs | `app/desks/` | `make mcp` |

`walk/` is the program `make demo` runs. It plays the story and prints each step. Your own APIs stay where they already run. You point `veto.yaml` at your OpenAPI.

## Harbor's APIs

Imagine these are already in production. `app/desks` hosts them for the demo.

| API | Address | What is there |
| --- | --- | --- |
| Orders | `127.0.0.1:18410` | Order 10482, Mara Ellison, a wool coat and two cedar trays, $556.00 |
| Customers | `127.0.0.1:18411` | Mara, `cus_mara` |
| Billing | `127.0.0.1:18412` | Invoice `inv_2291`, paid |

`warehouseId` is `wh_sfo_1`. It sits on the order.

Calls want `Authorization: Bearer demo-token`.

## Relationships

`app/relations.yaml`.

`customerId` on an order is the next call, `customers.get`. `invoiceId` is `invoices.get`. Add a line there to add a join. `warehouseId` has no line, so it stays a field.

## Semantics

`app/semantics.yaml`.

One sentence on `orders.delete`: this call waits for a person. One extra word, `scrap`. "Retire" is already built in. Search for either word and veto names `orders.delete`.

## CLI

```bash
make cli
```

Veto loads Harbor's contracts, prints the catalog and the joins, then the scripted agent hears "Delete order 10482". It selects `orders.delete` and stops before HTTP.

With the APIs up in another terminal (`make mcp`):

```bash
bin/veto validate --config app/veto.yaml
bin/veto preview --config app/veto.yaml --operation orders.delete --param id=10482
bin/veto approve <pending-id>
bin/veto eval --config app/veto.yaml --case app/cases
```

`veto approve` prints a different id from the pending one. That id runs once.

## MCP

Claude, Cursor, and ChatGPT use the same server.

```bash
make mcp
```

Paste `app/prompts/claude.md` as the instruction. The model gets three tools: `capabilities_search`, `capabilities_describe`, `capabilities_invoke`.

```json
{
  "mcpServers": {
    "veto": {
      "command": "bin/veto",
      "args": ["serve", "--config", "app/veto.yaml", "--stdio"],
      "env": {
        "DEMO_TOKEN": "demo-token"
      }
    }
  }
}
```

Use the absolute paths `make mcp` prints. Set `VETO_APPROVAL_NONCE_DIR` and `VETO_TOKEN_DIR` to the `.demo` directories it prints, so `veto approve` and the MCP process share the yes. Those directories belong to this demo.

## Agent

`make cli` is the agent with no API key. The scripted model asks to delete order 10482. Veto's policy stops the call. `app/cases` locks that same stop for `veto eval`.

A live agent is Claude, Cursor, or ChatGPT on `make mcp`, with `app/prompts/claude.md`. It reads Mara's order, follows the customer and the invoice, lets a cancel through, and stops when the delete asks for a person. You run `veto approve` and hand it the new id.

## What `make demo` prints

Mara's order, then her customer, then her invoice. The warehouse id stays a field. Jonas Adler's cancel reaches the orders API. Mara's delete waits until a person approves it, runs once, and the order is gone. The context pack holds the rules and the operation just described. The raw contract stays out of the model.
