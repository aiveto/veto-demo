Harbor sells home goods. Orders, customers, and billing are already up.

Start with `veto --help-json`. Search, describe, and invoke are the same three capabilities as MCP. `veto serve --json` serves JSON lines for skills and scripts. `--help-json` on those commands is the input schema. `--config app/veto.yaml` loads Harbor. Do not treat catalog flags as tool arguments.

Search before you describe. Describe one operation before you invoke it. Do not invent operation ids. Search hits include related ids and related_calls. Those lines are the next invoke.

A JSON object is the tool arguments. Words are a search shorthand.

```bash
veto search --config app/veto.yaml retire order 10482
veto search --config app/veto.yaml '{"query":"retire order 10482"}'
veto describe --config app/veto.yaml orders.get
veto invoke --config app/veto.yaml '{"operation_id":"orders.get","params":{"id":"10482"}}'
```

Mara Ellison asked to retire order 10482. Read the order first. The description names the next call for customerId and for invoiceId. Invoke those. warehouseId is on the order and is not a next call.

Deleting the order waits. If invoke returns confirmation_required, stop. Read the pending id to the person. Do not invent an approval and do not send the pending id back as approval_id. That returns pending_approval and does not run HTTP. They run `veto approve --config app/veto.yaml` in a shell that shares `VETO_APPROVAL_NONCE_DIR`, and give you a different id. Invoke orders.delete once with that id. sent is true when the request reached the transport. http is true when a response was received.

Cancelling an order is a different call. It does not wait.
