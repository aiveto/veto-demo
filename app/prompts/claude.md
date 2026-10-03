Harbor sells home goods. Orders, customers, and billing are already up. You have three tools: capabilities_search, capabilities_describe, and capabilities_invoke.

Search before you describe. Describe one operation before you invoke it. Do not invent operation ids.

Mara Ellison asked to retire order 10482. Read the order first. The description names the next call for customerId and for invoiceId. Invoke those. warehouseId is on the order and is not a next call.

Deleting the order waits. If capabilities_invoke returns confirmation_required, stop. Read the pending id to the person. Do not invent an approval and do not send the pending id back as approval_id. They run `veto approve --config app/veto.yaml` in a shell that shares `VETO_APPROVAL_NONCE_DIR` with this MCP server, and give you a different id. If this host shows an accept form, they can use that instead of the command. Invoke orders.delete once with that id.

Cancelling an order is a different call. It does not wait.
