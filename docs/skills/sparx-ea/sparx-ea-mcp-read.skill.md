---
name: sparx-ea-mcp-read
description: How to call the ArchiMate read tools of mcp-server-sparx-ea — navigate the tree, read an element, package or diagram, list the type vocabulary
whenToUse: when an agent needs to inspect an exported Sparx EA ArchiMate model (.xml) — the package/diagram/element tree, one element with its relationships, or the accepted ArchiMate types
tags:
  - skill/documentation/for-ai
  - concern/documentation
  - stack/mcp
---

# Goal
- Read an exported ArchiMate model: the whole tree, or one element / package / diagram, and the type vocabulary.

# Prerequisites
Register the server first — see [sparx-ea-mcp.skill.md](./sparx-ea-mcp.skill.md) and its [installation.md](./installation.md). All tools here take a `file` — a model the user exported from EA (`File → Export → Package to XMI`, XMI 2.1).

A **`ref`** is an EA id (`EAID_…` / `EAPK_…`), a braced GUID (`{F3C2209E-…}`), or a slash path from a root package (`Model/Motivation_Package/Goal1`). Get paths from `ea_model_tree`.

# Methods

## `ea_model_tree`

### Request
```
tools/call ea_model_tree { "file": "<path>" }
```

### Parameters
| Name | Type | Required |
| --- | --- | --- |
| `file` | string | yes |

### Return value
JSON `Node`, recursive:
```json
{ "kind": "package", "name": "Model", "id": "", "guid": "", "path": "",
  "children": [
    { "kind": "package", "name": "Motivation_Package", "id": "EAPK_…", "guid": "{…}",
      "path": "Model/Motivation_Package",
      "children": [
        { "kind": "diagram", "name": "Motivation_Diagram", "id": "EAID_…", "guid": "{…}", "path": "…" },
        { "kind": "element", "name": "Goal1", "id": "EAID_…", "guid": "{…}",
          "type": "ArchiMate.Goal", "path": "Model/Motivation_Package/Goal1" }
      ] } ] }
```
`kind` ∈ `package | diagram | element`. `type` is set on elements only;
non-ArchiMate content is `unknown:<X>`.

### Errors
`missing required argument: file`; a parse failure of the XML.

### Example
```
ea_model_tree { "file": "example/TestProject.xml" }
```

## `ea_element`

### Request
```
tools/call ea_element { "file": "<path>", "ref": "<ref>" }
```

### Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `file` | string | yes | |
| `ref` | string | yes | id, GUID or path of the element |

### Return value
```json
{ "id": "EAID_…", "guid": "{…}", "name": "Goal1", "type": "ArchiMate.Goal",
  "path": "Model/Motivation_Package/Goal1", "documentation": "…",
  "relations": [
    { "id": "EAID_…", "type": "ArchiMate.Realization", "name": "",
      "direction": "incoming", "otherName": "Requirement1", "otherId": "EAID_…",
      "otherType": "ArchiMate.Requirement" }
  ] }
```
`direction` is `outgoing` or `incoming` relative to this element.

### Errors
`sparx: no element for "<ref>"`.

### Example
```
ea_element { "file": "example/TestProject.xml", "ref": "Model/Motivation_Package/Goal1" }
```

## `ea_package`

### Request
```
tools/call ea_package { "file": "<path>", "ref": "<ref>" }
```

### Return value
```json
{ "id": "EAPK_…", "guid": "{…}", "name": "Motivation_Package",
  "path": "Model/Motivation_Package", "parent": "Model",
  "packages": [], "elements": ["Goal1", "Driver1"], "diagrams": ["Motivation_Diagram"] }
```
Child lists are names. `parent` is `""` for a root package.

### Errors
`sparx: no package for "<ref>"`.

## `ea_diagram`

### Request
```
tools/call ea_diagram { "file": "<path>", "ref": "<ref>" }
```

### Return value
```json
{ "id": "EAID_…", "guid": "{…}", "name": "Motivation_Diagram",
  "path": "…", "diagramType": "Logical",
  "objects": [ { "id": "EAID_…", "name": "Goal1", "type": "ArchiMate.Goal",
                 "seq": 1, "left": 460, "top": 250, "right": 560, "bottom": 320 } ],
  "links":   [ { "id": "EAID_…", "type": "ArchiMate.Realization", "name": "" } ] }
```
Coordinates are EA's: origin top-left, y downward.

### Errors
`sparx: no diagram for "<ref>"`.

## `ea_archimate_types`

### Request
```
tools/call ea_archimate_types {}
```
Takes **no arguments** and opens no file.

### Return value
```json
{ "elementTypes": ["ArchiMate.ApplicationComponent", "ArchiMate.Goal", "..."],
  "relationshipTypes": ["ArchiMate.Access", "ArchiMate.Realization", "..."] }
```

### Example
Call it before `ea_create_element` / `ea_create_relationship` to pick a valid `type`.

# Rule

## MUST
- Call `ea_model_tree` before any tool that needs a `ref` you do not already have.
- Read `type` from `ea_archimate_types` before passing one to an editing tool — an unknown type is refused there.
- Never pass an `output` argument to a read tool — reads do not write; that is `sparx-ea-mcp-edit.skill.md`.

## SHOULD
- Use `ea_element`'s `relations` (with `direction` and `otherName`) to understand how an element connects, rather than scanning the whole tree.
