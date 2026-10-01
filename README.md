# veto

Point veto at three contracts. One command shows what it does.

```bash
make demo
```

That needs a checkout of [veto](https://github.com/aiveto/veto) next to this directory, and Go 1.27.1. It builds veto, starts orders, customers, and billing, walks the catalog, and stops the desks.

## What you see

Mara Ellison's order is 10482: a wool coat and two cedar trays, $556.00. The customer is Mara. The invoice is paid. A warehouse id sits on the order.

The walk prints each of these.

**One catalog, three tools.** Twenty-five operations across the three contracts. The model gets `capabilities_search`, `capabilities_describe`, and `capabilities_invoke`. It does not get one tool per operation.

**Search knows the words.** "Retire order" and "scrap order" both name `orders.delete`. Retire is a built-in synonym. Scrap is one line in `semantics.yaml`, next to the sentence that this call waits for a person.

**A field is not a join.** Reading the order names the next call for `customerId` and for `invoiceId`. `warehouseId` stays a field. The joins are in `relations.yaml`.

**A write is not a delete.** Cancelling Jonas Adler's order reaches the desk. Deleting Mara's order does not. Preview, the scripted agent, and invoke all stop before HTTP. `veto approve` prints a different id. That id runs once. A second use is refused. The order is then gone.

**The model does not receive the raw spec.** The context pack holds the rules, a one-line index, and the operation just described. `veto eval` checks that stop, and that the pack does not contain the contract.

## Leave the desks up

```bash
make up
```

That prints the `veto serve` command for Claude or ChatGPT. Paste `prompts/claude.md` as the instruction. In another terminal, `make show` runs the same walk.

```json
{
  "mcpServers": {
    "veto": {
      "command": "bin/veto",
      "args": ["serve", "--config", "veto.yaml", "--stdio"],
      "env": {
        "DEMO_TOKEN": "demo-token"
      }
    }
  }
}
```

Use the absolute paths `make up` prints. Set `VETO_APPROVAL_NONCE_DIR` and `VETO_TOKEN_DIR` to the `.demo` directories it prints, so `veto approve` and the MCP process share the yes. Those directories are this demo's approvals, not yours.

The desks listen on `127.0.0.1:18410`, `:18411`, and `:18412`. API calls want `Authorization: Bearer demo-token`.

## Commands

`make` is `make demo`. `make help` lists the targets. `make vet` checks gofmt and `go vet`.
