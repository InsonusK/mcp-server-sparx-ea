---
name: sparx-ea-mcp
description: How to call the mcp-server-sparx-ea MCP server — read and edit an exported Sparx EA ArchiMate model
whenToUse: when an agent has the sparx-ea MCP server available and needs to inspect or change a Sparx Enterprise Architect ArchiMate model exported to XMI (.xml), or which child skill covers a specific capability
tags:
  - skill/documentation/for-ai
  - concern/documentation
  - stack/mcp
---

# Goal
- Give an agent everything needed to call the `mcp-server-sparx-ea` tools correctly: the ArchiMate read tools and the ArchiMate editing tools.

# Core Principle
- The server exposes **21 MCP tools** (names below). Every tool except `ea_new_model` takes a `file` path on the server's filesystem and never fetches anything over the network.
- All tools operate on a model the user exported from EA to XMI 2.1 (`File → Export → Package to XMI`), a `.xml` file. Editing tools also write a **new** `.xml` (`output`, must differ from `file`); the user re-imports it into EA. `ea_new_model` builds a model from scratch — it has no `file`, only `root` and `output`.
- Element and relationship types are ArchiMate 3.2 types (`ArchiMate.Goal`, `ArchiMate.Realization`), validated before any write. Call `ea_archimate_types` for the accepted list.
- Every failure is returned as an MCP **tool error** (`isError: true`, reason in the text), never a transport error. A service-level reason is prefixed `sparx:`.
- A tool result is JSON in a text content block. Editing tools return `{"result": <object>, "saved": "<output>"}`.

# Installation and access
See [installation.md](./installation.md) for building the server and registering it with an MCP client.

# Method-group skills

| Domain | Skill | Covers |
| --- | --- | --- |
| ArchiMate read | [sparx-ea-mcp-read.skill.md](./sparx-ea-mcp-read.skill.md) | `ea_model_tree`, `ea_element`, `ea_package`, `ea_diagram`, `ea_archimate_types` |
| ArchiMate edit | [sparx-ea-mcp-edit.skill.md](./sparx-ea-mcp-edit.skill.md) | `ea_new_model`, `ea_create_root_package`, `ea_set_root_name`, `ea_create_element`, `ea_update_element`, `ea_delete_element`, `ea_create_relationship`, `ea_delete_relationship`, `ea_create_package`, `ea_update_package`, `ea_copy_package`, `ea_delete_package`, `ea_create_diagram`, `ea_place_on_diagram`, `ea_move_on_diagram`, `ea_remove_from_diagram` |

# Rule

## MUST
- Address elements, packages and diagrams by an EA id (`EAID_…` / `EAPK_…`), a braced GUID (`{…}`), or a slash path from a root package (`Model/Motivation_Package/Goal1`). Call `ea_model_tree` first when you do not know the paths.
- For any editing tool, pass an `output` path that is different from `file` — the tool refuses `output == file` (`output must be a different path from file`).
- Chain edits by feeding each tool's `output` to the next tool's `file`. Never point two edits at the same `output` expecting them to accumulate — the second overwrites the first.
- Treat a tool result with `isError: true` as data: read the reason, fix the arguments or report it. Never retry the identical call.
- Never add a new tool's documentation to this root skill once it belongs to a domain — put it in the matching child skill, or discovery becomes noisy.

## SHOULD
- Read `sparx-ea-mcp-read.skill.md` before `sparx-ea-mcp-edit.skill.md` — every edit needs a `ref` you get from a read.
- Tell the user that an editing tool's `output` file must be imported into EA (`File → Import → Package from XMI`) for the change to take effect.

# Check list
- [ ] `installation.md` is the only place install/registration is documented; every child skill links to it.
- [ ] Each child skill's `whenToUse` names its own tools and does not overlap the sibling.
- [ ] `# Method-group skills` lists both child skills and every tool is in exactly one of them.
