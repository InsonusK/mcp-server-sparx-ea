---
name: sparx-ea-mcp
description: How to call the mcp-server-sparx-ea MCP server — read a Sparx EA project with SQL, and read/edit an exported ArchiMate model
whenToUse: when an agent has the sparx-ea MCP server available and needs to inspect a Sparx Enterprise Architect project (.eapx) or an exported ArchiMate model (.xml), or which child skill covers a specific capability
tags:
  - skill/documentation/for-ai
  - concern/documentation
  - stack/mcp
---

# Goal
- Give an agent everything needed to call the `mcp-server-sparx-ea` tools correctly: the SQL query tool, the ArchiMate read tools, and the ArchiMate editing tools.

# Core Principle
- The server exposes **18 MCP tools** (names below). Every tool takes a `file` path on the server's filesystem and never fetches anything over the network.
- Two independent surfaces:
  - **`ea_query`** — read-only SQL against the `.eapx` / `.eap` database. Nothing else touches the `.eapx`.
  - **`ea_*` ArchiMate tools** — operate on a model the user exported from EA to XMI 2.1 (`File → Export → Package to XMI`), a `.xml` file. Editing tools write a **new** `.xml` (`output`, must differ from `file`); the user re-imports it into EA.
- Element and relationship types are ArchiMate 3.2 types (`ArchiMate.Goal`, `ArchiMate.Realization`), validated before any write. Call `ea_archimate_types` for the accepted list.
- Every failure is returned as an MCP **tool error** (`isError: true`, reason in the text), never a transport error. A service-level reason is prefixed `sparx:`.
- A tool result is JSON in a text content block. Editing tools return `{"result": <object>, "saved": "<output>"}`.

# Installation and access
See [installation.md](./installation.md) for building the server and registering it with an MCP client.

# Method-group skills

| Domain | Skill | Covers |
| --- | --- | --- |
| SQL | [sparx-ea-mcp-sql.skill.md](./sparx-ea-mcp-sql.skill.md) | `ea_query` — read-only SQL against the `.eapx` database |
| ArchiMate read | [sparx-ea-mcp-read.skill.md](./sparx-ea-mcp-read.skill.md) | `ea_model_tree`, `ea_element`, `ea_package`, `ea_diagram`, `ea_archimate_types` |
| ArchiMate edit | [sparx-ea-mcp-edit.skill.md](./sparx-ea-mcp-edit.skill.md) | `ea_create_element`, `ea_update_element`, `ea_delete_element`, `ea_create_relationship`, `ea_delete_relationship`, `ea_create_package`, `ea_update_package`, `ea_copy_package`, `ea_delete_package`, `ea_place_on_diagram`, `ea_move_on_diagram`, `ea_remove_from_diagram` |

# Rule

## MUST
- Address elements, packages and diagrams by an EA id (`EAID_…` / `EAPK_…`), a braced GUID (`{…}`), or a slash path from a root package (`Model/Motivation_Package/Goal1`). Call `ea_model_tree` first when you do not know the paths.
- For any editing tool, pass an `output` path that is different from `file` and from the `.eapx` — the tool refuses `output == file` (`output must be a different path from file`).
- Chain edits by feeding each tool's `output` to the next tool's `file`. Never point two edits at the same `output` expecting them to accumulate — the second overwrites the first.
- Treat a tool result with `isError: true` as data: read the reason, fix the arguments or report it. Never retry the identical call.
- Never use `ea_query` to change data — the `.eapx` is opened read-only and write statements are rejected (`only read-only SELECT statements are supported`). Use the ArchiMate editing tools on an XMI export instead.
- Never add a new tool's documentation to this root skill once it belongs to a domain — put it in the matching child skill, or discovery becomes noisy.

## SHOULD
- Read `sparx-ea-mcp-read.skill.md` before `sparx-ea-mcp-edit.skill.md` — every edit needs a `ref` you get from a read.
- Tell the user that an editing tool's `output` file must be imported into EA (`File → Import → Package from XMI`) for the change to take effect.

## MAY
- Skip `ea_query` entirely if the ArchiMate tools already expose what you need — it is for raw table data (attributes, operations, tagged values, geometry) only.

# Check list
- [ ] `installation.md` is the only place install/registration is documented; every child skill links to it.
- [ ] Each child skill's `whenToUse` names its own tools and does not overlap the siblings.
- [ ] `# Method-group skills` lists all three child skills and every tool is in exactly one of them.
