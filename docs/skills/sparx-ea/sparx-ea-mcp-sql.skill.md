---
name: sparx-ea-mcp-sql
description: How to call ea_query — read-only SQL against a Sparx EA .eapx / .eap database
whenToUse: when an agent needs raw table data from a Sparx EA project file (attributes, operations, tagged values, connector geometry, audit columns) that the ArchiMate tools do not expose
tags:
  - skill/documentation/for-ai
  - concern/documentation
  - stack/mcp
---

# Goal
- Call `ea_query` to run one `SELECT` against a Sparx EA project database and get the rows as JSON.

# Prerequisites
Register the server first — see [sparx-ea-mcp.skill.md](./sparx-ea-mcp.skill.md) and its [installation.md](./installation.md).

# Methods

## `ea_query`

### Request

```
tools/call ea_query { "file": "<path>", "sql": "<SELECT ...>" }
```

### Parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `file` | string | yes | Path to the `.eapx` / `.eap` file on the server. |
| `sql` | string | yes | One read-only `SELECT`. `INSERT/UPDATE/DELETE/DROP/…` are rejected. |

### Return value

Text content block containing JSON:

```json
{ "columns": ["Object_ID", "Name"], "rows": [["2", "Stakeholder1"]], "rowCount": 1 }
```

- `columns: string[]`, `rows: string[][]` (always an array, never `null`), `rowCount: number`.
- All cell values are strings; text is decoded to UTF-8.

### Errors (tool errors, `isError: true`)

| Text contains | Cause | Handling |
| --- | --- | --- |
| `missing required argument: file` / `... sql` | argument omitted | add it |
| `file not found` / `Unable to locate database` | bad path / not a JET DB | fix `file` |
| `only read-only SELECT statements are supported` | write / DDL statement | use the ArchiMate editing tools on an XMI export instead |
| mdbtools error text | SQL the backend cannot run | rewrite; see "Supported SQL" |

### Supported SQL

```
SELECT <cols> FROM <table> [WHERE <col> = / <> / LIKE <val> [AND / OR ...]]
```

`ORDER BY`, `LIMIT`, `IN (...)`, `JOIN`, `count(*)` with `WHERE` — **not
supported**. Rows come back in storage order; select one row with
`WHERE <key> = <n>` or compare multi-row results as sets.

### Example

```
ea_query { "file": "example/TestProject.eapx",
           "sql": "select Object_ID, Name, Stereotype from t_object where Object_Type = 'Class'" }
```

Expected output:

```json
{
  "columns": ["Object_ID", "Name", "Stereotype"],
  "rows": [["2", "Stakeholder1", "ArchiMate_Stakeholder"], ["4", "Goal1", "ArchiMate_Goal"]],
  "rowCount": 2
}
```

### Useful tables

`t_object` (elements: `Object_ID`, `Name`, `Object_Type`, `Stereotype`, `Note`,
`Package_ID`), `t_connector` (`Connector_ID`, `Connector_Type`,
`Start_Object_ID`, `End_Object_ID`, `Stereotype`), `t_package`
(`Package_ID`, `Name`, `Parent_ID`), `t_diagram`, `t_diagramobjects`
(`RectLeft/Top/Right/Bottom`), `t_attribute`, `t_operation`,
`t_objectproperties` (tagged values).

# Rule

## MUST
- Pass a single `SELECT`. Never send a write or DDL statement — it is rejected, and the `.eapx` is read-only through this tool.
- Compare multi-row results as sets — there is no `ORDER BY`.
- Never document or call an ArchiMate tool from this skill — those belong to `sparx-ea-mcp-read.skill.md` / `sparx-ea-mcp-edit.skill.md`.

## SHOULD
- Prefer the ArchiMate read tools for model structure (packages, elements, relationships, diagrams); reach for `ea_query` only for columns they do not return.
