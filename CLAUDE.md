# veto-demo

Veto is the product. Harbor is the sample merchant.

Veto is the sibling checkout `../veto`. `make` builds that checkout to `bin/veto`. Do not copy veto's source into `app/`.

`app/` is Harbor's three APIs, already hosted for the demo, plus the files a person edits: `relations.yaml`, `semantics.yaml`, `veto.yaml`, `prompts/claude.md`, and `cases/`. `walk/` is the program `make demo` runs. A user's own APIs stay in their own repo.

`make demo` plays the story. `make cli` runs the CLI and the scripted agent. `make mcp` leaves Harbor's APIs listening for Claude, Cursor, or ChatGPT.

Approvals and tokens stay under `.demo/`. Do not point them at a real veto approval directory. Leave `VETO_APPROVAL_SECRET` unset.

Do not add a second product name. Harbor is the sample. Do not register one MCP tool per operation. The tools are `capabilities_search`, `capabilities_describe`, and `capabilities_invoke`.

`app/prompts/claude.md` is the instruction for a model using those tools. This file is for an agent editing the repository.
