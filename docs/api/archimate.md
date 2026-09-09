# ArchiMate tools

Read and edit an [ArchiMate](../glossary/archimate.md) model that the user
exported from Sparx EA to [XMI](../glossary/xmi.md). Types are ArchiMate types
(`ArchiMate.Goal`), not `uml:Class` + stereotype; every edit is validated
against the ArchiMate 3.2 vocabulary and relationship rules before it is
written.

## Setup

Install and register the server first — see the `README.md` and
[docs/installation.md](../installation.md). Read tools take a `file` (the
exported `.xml`). Editing tools also take an `output` and write the changed
copy there; `output` must differ from `file`. (`ea_new_model` is the exception —
it builds a model from scratch, so it takes `root` + `output` and no `file`.)
Read [the editing workflow](../workflow.md) before making changes.

## Common conventions

- **`ref`** — an element, package or diagram is addressed by its EA id
  (`EAID_…` / `EAPK_…`), its braced GUID (`{F3C2209E-…}`), or a slash path from
  a root package (`Model/Motivation_Package/Goal1`).
- **Editing tools** return `{"result": <object>, "saved": "<output path>"}`.
- **Errors** are MCP *tool errors* (the call is not a transport failure); the
  text starts with `sparx:` for a service-level reason.
- A missing required argument returns `missing required argument: <name>`.

---

## Read tools

### `ea_model_tree`

The whole model as a tree of packages, diagrams and elements with ids — like
the EA project browser.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file` | string | yes | the exported `.xml` model |

**Returns** a `Node`: `{kind, name, id, guid, type, path, children}`. `kind` is
`package`, `diagram` or `element`; `type` is the ArchiMate type for elements
(`unknown:<X>` for non-ArchiMate content).

```
ea_model_tree  file=example/TestProject.xml
```

### `ea_element`

One element by `ref`: its ArchiMate type, note and every relationship.

| Arg | Type | Required |
| --- | --- | --- |
| `file` | string | yes |
| `ref` | string | yes |

**Returns** `{id, guid, name, type, path, documentation, relations[]}`. Each
`relation` is `{id, type, name, direction, otherName, otherId, otherType}` —
`direction` is `outgoing` or `incoming` relative to this element.

**Errors** — `sparx: no element for "<ref>"`.

```
ea_element  file=example/TestProject.xml  ref="Model/Motivation_Package/Goal1"
```

### `ea_package`

One package by `ref`: its parent and immediate children.

| Arg | Type | Required |
| --- | --- | --- |
| `file` | string | yes |
| `ref` | string | yes |

**Returns** `{id, guid, name, path, parent, packages[], elements[], diagrams[]}`
— the child lists are names.

### `ea_diagram`

One diagram by `ref`: its type and what is placed on it.

| Arg | Type | Required |
| --- | --- | --- |
| `file` | string | yes |
| `ref` | string | yes |

**Returns** `{id, guid, name, path, diagramType, objects[], links[]}`. Each
`object` is `{id, name, type, seq, left, top, right, bottom}` (EA coordinates:
origin top-left); each `link` is `{id, type, name}`.

### `ea_archimate_types`

The type names this server accepts. Takes **no arguments** and opens no file.

**Returns** `{"elementTypes": ["ArchiMate.ApplicationComponent", …], "relationshipTypes": ["ArchiMate.Access", …]}`.

```
ea_archimate_types
```

---

## Model tools

### `ea_new_model`

Build a model from scratch — one EA root package named `root`, with the ArchiMate3
profile embedded so ArchiMate elements are recognised on import. **No `file`.**

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `root` | string | yes | the single root package name |
| `output` | string | yes | where to write the new model |

**Returns** `{"result": <tree>, "saved": "<output>"}`. Chain the other tools with
`file` = this `output`.

### `ea_create_root_package`

Add a new top-level package alongside the model root.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | `output` ≠ `file` |
| `name` | string | yes | the new package name |

**Returns** `{"result": <PackageInfo>, "saved": "<output>"}`.

### `ea_set_root_name`

Rename the single EA root package. With `fresh_identity` it also regenerates every
GUID, so the saved copy imports into EA as an independent package next to the
original.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | `output` ≠ `file` |
| `name` | string | yes | the new root name |
| `fresh_identity` | bool | no | also regenerate every GUID (default `false`) |

**Returns** `{"result": {"rootName": "<new name>"}, "saved": "<output>"}`.

---

## Element tools

### `ea_create_element`

Add an ArchiMate element to a package.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file` | string | yes | the exported model |
| `output` | string | yes | where to write the edited copy (≠ `file`) |
| `parent` | string | yes | the package to add it to (`ref`) |
| `type` | string | yes | ArchiMate element type, e.g. `ArchiMate.Requirement` (`ea_archimate_types` lists them) |
| `name` | string | yes | element name |
| `note` | string | no | documentation |

**Returns** `{"result": <ElementInfo>, "saved": "<output>"}`.

**Errors**

| Cause | Message contains |
| --- | --- |
| `type` not in the ArchiMate vocabulary | `is not a known ArchiMate element type` |
| `parent` does not resolve | `no package for` |
| the package already has an element with that name | `already has an element named` |
| `output` equals `file` | `output must be a different path from file` |

```
ea_create_element  file=model.xml  output=/tmp/edited.xml
                   parent="Model/Motivation_Package"  type=ArchiMate.Goal  name="Reduce cost"
```

### `ea_update_element`

Rename an element, change its note and/or move it to another package. Only the
fields you pass are changed; passing none just returns the current element.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | |
| `ref` | string | yes | the element |
| `name` | string | no | new name |
| `note` | string | no | new documentation (pass an empty string to clear) |
| `parent` | string | no | move it into this package |

**Errors** — `no element for`, `new name is required` (blank `name`),
`no package for` (bad `parent`), `already has an element named` (name clash on
move).

```
ea_update_element  file=model.xml  output=/tmp/edited.xml
                   ref="Model/Motivation_Package/Value1"  name="CustomerValue"  note="value to the customer"
```

### `ea_delete_element`

Remove an element **and every relationship attached to it**, and its placements
on any diagram.

| Arg | Type | Required |
| --- | --- | --- |
| `file`, `output` | string | yes |
| `ref` | string | yes |

**Returns** `{"result": {"deleted": true}, "saved": "<output>"}`.
**Errors** — `no element for`.

---

## Relationship tools

### `ea_create_relationship`

Add an ArchiMate relationship between two elements.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | |
| `source` | string | yes | source element (`ref`) |
| `target` | string | yes | target element (`ref`) |
| `type` | string | yes | ArchiMate relationship type, e.g. `ArchiMate.Realization` |
| `name` | string | no | connector name |
| `note` | string | no | documentation |

**Returns** `{"result": <Relation>, "saved": "<output>"}`.

**Errors**

| Cause | Message contains |
| --- | --- |
| `type` not a known relationship | `is not a known ArchiMate relationship type` |
| `source` / `target` do not resolve | `no source element for` / `no target element for` |
| ArchiMate does not permit it between those types | `is not allowed:` + a reason |
| a `(type, source, target)` relationship already exists | `already exists` |

```
ea_create_relationship  file=model.xml  output=/tmp/edited.xml
                        source="Model/Motivation_Package/Requirement1"
                        target="Model/Motivation_Package/Goal1"  type=ArchiMate.Realization
```

### `ea_delete_relationship`

Remove a relationship by id, **or** every relationship between two elements.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | |
| `id` | string | no | the relationship id or GUID |
| `source` | string | no | with `target`: delete every relationship source→target |
| `target` | string | no | with `source` |

Pass either `id` **or** both `source` and `target`.

**Returns** — by id: `{"result": {"deleted": true}, "saved": …}`; by endpoints:
`{"result": {"deleted": <count>}, "saved": …}`.

**Errors** — `pass either 'id' or both 'source' and 'target'`,
`no relationship "<id>"`, `no relationship from <A> to <B>`.

---

## Package tools

### `ea_create_package`

Add a package inside another package.

| Arg | Type | Required |
| --- | --- | --- |
| `file`, `output` | string | yes |
| `parent` | string | yes |
| `name` | string | yes |

**Errors** — `no package for` (bad `parent`), `already has a sub-package named`.

### `ea_update_package`

Rename a package and/or move it to another parent.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | |
| `ref` | string | yes | the package |
| `name` | string | no | new name |
| `parent` | string | no | move it into this package |

**Errors** — `no package for`, `new name is required`,
`cannot move a package into its own descendant`, `already has a sub-package named`.

### `ea_copy_package`

Deep-copy a package (with a **fresh identity**, so nothing collides) into
another package.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | |
| `ref` | string | yes | the package to copy |
| `dest` | string | yes | the package to copy it into |

References that point *out* of the copied subtree (a relationship end or a
diagram object that names an element in another package) are left as-is.

**Errors** — `no package for`,
`cannot copy a package into itself or its own descendant`,
`already has a sub-package named`.

### `ea_delete_package`

Remove a package.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | |
| `ref` | string | yes | the package |
| `cascade` | boolean | no | also delete every sub-package, element and diagram inside it (default `false`) |

Without `cascade` a non-empty package is refused. An EA root package is never
deleted.

**Returns** `{"result": {"removed": <count>}, "saved": …}` — the number of
model objects removed.

**Errors** — `no package for`,
`is an EA root package; rename it instead of deleting`,
`is not empty; pass cascadeDelete to remove its contents`.

---

## Diagram tools

### `ea_place_on_diagram` / `ea_move_on_diagram`

Place an element on a diagram at a rectangle, or move one that is already there.

| Arg | Type | Required | Description |
| --- | --- | --- | --- |
| `file`, `output` | string | yes | |
| `diagram` | string | yes | the diagram (`ref`) |
| `element` | string | yes | the element (`ref`) |
| `left`, `top`, `right`, `bottom` | number | yes | rectangle in EA coordinates (origin top-left, y down); `right > left`, `bottom > top` |

**Returns** `{"result": <DiagramInfo>, "saved": …}` — the diagram after the change.

**Errors** — `no diagram for`, `no element for`, `is already on diagram`
(place), `is not on diagram` (move), `invalid rectangle`.

```
ea_place_on_diagram  file=model.xml  output=/tmp/edited.xml
                     diagram="Model/Motivation_Package/Motivation_Diagram"
                     element="Model/Motivation_Package/Goal1"  left=100 top=100 right=200 bottom=170
```

### `ea_remove_from_diagram`

Remove an element's placement from a diagram (the element stays in the model).

| Arg | Type | Required |
| --- | --- | --- |
| `file`, `output` | string | yes |
| `diagram` | string | yes |
| `element` | string | yes |

**Errors** — `no diagram for`, `no element for`, `is not on diagram`.

## See also

- [The editing workflow](../workflow.md) — export, chain edits, re-import.
