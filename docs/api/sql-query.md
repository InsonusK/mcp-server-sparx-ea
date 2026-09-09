# SQL query

Read-only SQL against a Sparx EA project database (`.eapx` / `.eap`). Use this
when you need raw table data that the ArchiMate tools do not expose — attributes,
operations, tagged values, connector geometry, audit columns.

## Setup

Install and register the server first — see the `README.md` and
[docs/installation.md](../installation.md). This tool needs the `.eapx` file on
the server's filesystem.

## Tools

### `ea_query`

Runs one `SELECT` statement and returns the rows as JSON.

**Arguments**

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `file` | string | yes | Path to the `.eapx` (or `.eap`) file on the server. |
| `sql` | string | yes | A single read-only `SELECT` statement. |

**Returns** — a text block containing JSON:

```json
{ "columns": ["Object_ID", "Name"], "rows": [["2", "Stakeholder1"]], "rowCount": 1 }
```

`rows` is always an array (never `null`) — an empty result is `{"columns":[...],"rows":[],"rowCount":0}`.
All values are strings; text is decoded to UTF-8 (JET4 stores it as UCS-2LE).

**Errors** — returned as an MCP *tool error* (not a transport error), text
contains the reason:

| Cause | Message contains |
| --- | --- |
| `file` or `sql` missing | `missing required argument: file` / `... sql` |
| File not found / not a database | `file not found` / `Unable to locate database` |
| A write statement (`INSERT`, `UPDATE`, `DELETE`, `DROP`, …) | `only read-only SELECT statements are supported` |
| SQL the backend cannot run | the mdbtools error text |

**Supported SQL** — a thin pass-through to the mdbtools engine:

```
SELECT <cols> FROM <table> [WHERE <col> = / <> / LIKE <val> [AND / OR ...]]
```

`ORDER BY`, `LIMIT`, `IN (...)`, `JOIN`, and `count(*)` with a `WHERE` clause are
**not** supported. Rows come back in storage order — compare multi-row results
as sets, or select exactly one row with `WHERE <key> = <n>`.

**Example**

```
ea_query  file=example/TestProject.eapx
          sql="select Object_ID, Name, Object_Type from t_object where Object_Type = 'Class'"
```

```json
{
  "columns": ["Object_ID", "Name", "Object_Type"],
  "rows": [["2", "Stakeholder1", "Class"], ["3", "Driver1", "Class"]],
  "rowCount": 2
}
```

**Useful tables**

| Table | Holds |
| --- | --- |
| `t_object` | Elements (`Object_ID`, `Name`, `Object_Type`, `Stereotype`, `Note`, `Package_ID`) |
| `t_connector` | Relationships (`Connector_ID`, `Connector_Type`, `Start_Object_ID`, `End_Object_ID`, `Stereotype`) |
| `t_package` | Packages (`Package_ID`, `Name`, `Parent_ID`) |
| `t_diagram` | Diagrams |
| `t_diagramobjects` | Element placements on diagrams (`RectLeft`, `RectTop`, …) |
| `t_attribute`, `t_operation` | Attributes and operations of an element |
| `t_objectproperties` | Tagged values |

## See also

- [ArchiMate tools](archimate.md) — the higher-level model API.
- [The editing workflow](../workflow.md) — how to change the model (SQL is read-only).
