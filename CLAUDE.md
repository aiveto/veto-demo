# veto-demo

This repository is the demo. The veto module is the sibling checkout `../veto`.

`make demo` builds veto from that checkout, starts orders, customers, and billing, runs the walk, and stops the desks. `make up` leaves the desks listening. `make vet` is gofmt and `go vet`.

Approvals and tokens stay under `.demo/`. Do not point them at a real veto approval directory. Leave `VETO_APPROVAL_SECRET` unset.

The name on the page is veto. Do not add a second product name. Do not register one MCP tool per operation. The tools are `capabilities_search`, `capabilities_describe`, and `capabilities_invoke`.

`prompts/claude.md` is the instruction for a model using those tools. This file is for an agent editing the repository.

```
cmd/apis/    the three desks
cmd/show/    the walk
apis/        in-memory orders, customers, and billing
contracts/   OpenAPI
cases/       veto eval
```
