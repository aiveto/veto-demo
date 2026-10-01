# veto-demo

This repository is a sample application in front of veto, plus a walk.

Veto is the sibling checkout `../veto`. `make` builds that checkout to `bin/veto`. Do not copy veto's source into `app/`.

`app/` is the sample of what a user writes: contracts, services, `veto.yaml`, semantics, relations, the model instruction, and eval cases. `walk/` drives the demonstration. Do not treat `walk/` as part of the application.

`make demo` builds veto, starts the services, runs the walk, and stops them. `make up` leaves the services listening. `make vet` is gofmt and `go vet`.

Approvals and tokens stay under `.demo/`. Do not point them at a real veto approval directory. Leave `VETO_APPROVAL_SECRET` unset.

The name on the page is veto. Do not add a second product name. Do not register one MCP tool per operation. The tools are `capabilities_search`, `capabilities_describe`, and `capabilities_invoke`.

`app/prompts/claude.md` is the instruction for a model using those tools. This file is for an agent editing the repository.
