# Harbor

Harbor is a home-goods desk: orders, customers, and billing. Twenty-five operations. The agent gets three tools.

`make demo` reads Mara Ellison's order, follows the customer and the invoice because the relation says so, lets a cancel through, and holds the delete until a person approves it. The approved id runs once.

```bash
make demo
```

That needs a checkout of [veto](https://github.com/aiveto/veto) next to this directory, Go 1.27.1, and nothing else. The desks listen on `127.0.0.1:18410`, `:18411`, and `:18412`. The token is `harbor-demo-token`. Approvals stay in `.harbor/` and are not your real veto approvals.

## Leave the desks up

```bash
make up
```

That prints the `veto serve` command for Claude or ChatGPT. Paste `prompts/claude.md` as the instruction. In another terminal, `make show` runs the same walk.

The server entry is the three tools, not one tool per operation:

```json
{
  "mcpServers": {
    "harbor": {
      "command": "bin/veto",
      "args": ["serve", "--config", "veto.yaml", "--stdio"],
      "env": {
        "HARBOR_TOKEN": "harbor-demo-token"
      }
    }
  }
}
```

Use the absolute paths `make up` prints. Set `VETO_APPROVAL_NONCE_DIR` and `VETO_TOKEN_DIR` to the `.harbor` directories it prints, so `veto approve` and the MCP process share the yes.

## What you are seeing

Search for "retire order" or "scrap order" names `orders.delete`. Retire is a built-in synonym. Scrap is one line in `semantics.yaml`.

`orders.get` returns `customerId`, `invoiceId`, and `warehouseId`. Only the first two are joins. They are in `relations.yaml`. The field name does not create the call.

Cancel is a write. It reaches the desk. Delete waits. Preview, the scripted agent, and `capabilities_invoke` all stop before HTTP. `veto approve` prints a different id. That id runs once.

`veto eval` checks the same stop, and that the context pack does not contain the raw spec.
