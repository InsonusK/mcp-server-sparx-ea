---
name: sparx-ea-mcp-edit
description: How to call the ArchiMate editing tools of mcp-server-sparx-ea — create/update/delete elements, relationships and packages, and place elements on diagrams, writing an importable XMI copy
whenToUse: when an agent needs to change an exported Sparx EA ArchiMate model (.xml) — add or edit elements, relationships or packages, or place elements on a diagram — and produce a file the user re-imports into EA
tags:
  - skill/documentation/for-ai
  - concern/documentation
  - stack/mcp
---

# Goal
- Edit an exported ArchiMate model with validated operations and write a new `.xml` the user re-imports into EA.

# Prerequisites
Register the server first — see [sparx-ea-mcp.skill.md](./sparx-ea-mcp.skill.md) and its [installation.md](./installation.md). Get every `ref` and every `type` from the read tools first — see [sparx-ea-mcp-read.skill.md](./sparx-ea-mcp-read.skill.md).

# Core conventions
- Every tool here takes **`file`** (the input model) and **`output`** (where to write the edited copy). `output` must be a different path from `file`; the tool refuses `output == file`.
- **Chain edits**: pass one tool's `output` as the next tool's `file`. Two tools writing the same `output` do not accumulate — the second overwrites.
- Result shape: `{"result": <object>, "saved": "<output>"}`. `<object>` is the affected element / relationship / package / diagram after the change.
- A `ref` is an id, a braced GUID, or a slash path (`Model/Pkg/Elem`).
- Missing required arg → `missing required argument: <name>`. Service reason → `sparx: …`.
- After the last edit, tell the user to import the final `output` into EA: `File → Import → Package from XMI`.

# Methods

## `ea_create_element`
```
tools/call ea_create_element
  { "file": "<in>", "output": "<out>", "parent": "<ref>",
    "type": "ArchiMate.Goal", "name": "<name>", "note": "<optional>" }
```
| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | `output ≠ file` |
| `parent` | string | yes | package `ref` |
| `type` | string | yes | ArchiMate element type — `ea_archimate_types` lists them |
| `name` | string | yes | |
| `note` | string | no | documentation |

**Returns** `{"result": <ElementInfo>, "saved": "<out>"}`.
**Errors** — `is not a known ArchiMate element type`, `no package for`, `already has an element named`, `output must be a different path from file`.

## `ea_update_element`
```
tools/call ea_update_element
  { "file": "<in>", "output": "<out>", "ref": "<ref>",
    "name": "<optional>", "note": "<optional>", "parent": "<optional ref>" }
```
Applies only the fields you pass (rename, then re-note, then move). Passing none returns the current element.
**Errors** — `no element for`, `new name is required`, `no package for`, `already has an element named`.

## `ea_delete_element`
```
tools/call ea_delete_element { "file": "<in>", "output": "<out>", "ref": "<ref>" }
```
Removes the element, every relationship attached to it, and its diagram placements.
**Returns** `{"result": {"deleted": true}, "saved": "<out>"}`. **Errors** — `no element for`.

## `ea_create_relationship`
```
tools/call ea_create_relationship
  { "file": "<in>", "output": "<out>", "source": "<ref>", "target": "<ref>",
    "type": "ArchiMate.Realization", "name": "<optional>", "note": "<optional>" }
```
| Name | Type | Required |
| --- | --- | --- |
| `file`, `output`, `source`, `target`, `type` | string | yes |
| `name`, `note` | string | no |

**Returns** `{"result": <Relation>, "saved": "<out>"}`.
**Errors** — `is not a known ArchiMate relationship type`; `no source element for` / `no target element for`; `is not allowed:` + reason (ArchiMate forbids it between those types); `already exists` (a `(type, source, target)` relationship is already there — it is a unique key).

## `ea_delete_relationship`
```
tools/call ea_delete_relationship { "file": "<in>", "output": "<out>", "id": "<rel id>" }
# or
tools/call ea_delete_relationship { "file": "<in>", "output": "<out>", "source": "<ref>", "target": "<ref>" }
```
Pass **either** `id` **or** both `source` and `target` (deletes every relationship source→target).
**Returns** — by id `{"result": {"deleted": true}, …}`; by endpoints `{"result": {"deleted": <count>}, …}`.
**Errors** — `pass either 'id' or both 'source' and 'target'`, `no relationship "<id>"`, `no relationship from <A> to <B>`.

## `ea_create_package`
```
tools/call ea_create_package { "file": "<in>", "output": "<out>", "parent": "<ref>", "name": "<name>" }
```
**Errors** — `no package for`, `already has a sub-package named`.

## `ea_update_package`
```
tools/call ea_update_package
  { "file": "<in>", "output": "<out>", "ref": "<ref>", "name": "<optional>", "parent": "<optional ref>" }
```
Rename then move. **Errors** — `no package for`, `new name is required`, `cannot move a package into its own descendant`, `already has a sub-package named`.

## `ea_copy_package`
```
tools/call ea_copy_package { "file": "<in>", "output": "<out>", "ref": "<ref>", "dest": "<ref>" }
```
Deep-copies the package (fresh GUIDs) into `dest`. References that point out of the copied subtree are left as-is.
**Errors** — `no package for`, `cannot copy a package into itself or its own descendant`, `already has a sub-package named`.

## `ea_delete_package`
```
tools/call ea_delete_package { "file": "<in>", "output": "<out>", "ref": "<ref>", "cascade": false }
```
| Name | Type | Required | Default |
| --- | --- | --- | --- |
| `file`, `output`, `ref` | string | yes | |
| `cascade` | boolean | no | `false` |

Without `cascade`, a non-empty package is refused. A root package is never deleted.
**Returns** `{"result": {"removed": <count>}, "saved": "<out>"}`.
**Errors** — `no package for`, `is an EA root package; rename it instead of deleting`, `is not empty; pass cascadeDelete to remove its contents`.

## `ea_place_on_diagram` / `ea_move_on_diagram`
```
tools/call ea_place_on_diagram
  { "file": "<in>", "output": "<out>", "diagram": "<ref>", "element": "<ref>",
    "left": 100, "top": 100, "right": 200, "bottom": 170 }
```
| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| `file`, `output`, `diagram`, `element` | string | yes | |
| `left`, `top`, `right`, `bottom` | number | yes | EA coordinates (origin top-left, y down); `right > left`, `bottom > top` |

**Returns** `{"result": <DiagramInfo>, "saved": "<out>"}`.
**Errors** — `no diagram for`, `no element for`, `is already on diagram` (place), `is not on diagram` (move), `invalid rectangle`.

## `ea_remove_from_diagram`
```
tools/call ea_remove_from_diagram { "file": "<in>", "output": "<out>", "diagram": "<ref>", "element": "<ref>" }
```
Removes the placement; the element stays in the model. **Errors** — `no diagram for`, `no element for`, `is not on diagram`.

# Worked example: add a goal and realize it

```
1. ea_model_tree     { "file": "model.xml" }
     → find "Model/Motivation_Package" and an existing "Requirement1"
2. ea_create_element { "file": "model.xml", "output": "step1.xml",
                       "parent": "Model/Motivation_Package", "type": "ArchiMate.Goal", "name": "Reduce cost" }
     → { "result": { "path": "Model/Motivation_Package/Reduce cost", ... }, "saved": "step1.xml" }
3. ea_create_relationship { "file": "step1.xml", "output": "step2.xml",
                            "source": "Model/Motivation_Package/Requirement1",
                            "target": "Model/Motivation_Package/Reduce cost",
                            "type": "ArchiMate.Realization" }
     → { "result": { "type": "ArchiMate.Realization", ... }, "saved": "step2.xml" }
4. Tell the user: import step2.xml into EA (File → Import → Package from XMI).
```

# Rule

## MUST
- Pass an `output` different from `file` on every call — `output == file` is refused, and never edit a `.eapx` this way.
- Chain edits through `output → file`. Never send two edits to the same `output` expecting them to combine.
- Get `ref`s and `type`s from the read tools (`sparx-ea-mcp-read.skill.md`) before editing — an unresolved `ref` or unknown `type` is a tool error, not a silent no-op.
- Tell the user the final `output` file must be imported into EA for the change to land; the tools never touch the live project.
- Never document or call `ea_query` or a read tool from this skill — they belong to their own skills.

## SHOULD
- Check `ea_archimate_types` output when a `type` is refused, rather than guessing another name.
- Prefer `ea_delete_relationship` by `id` (from `ea_element`'s `relations`) when you know exactly which one to remove.
