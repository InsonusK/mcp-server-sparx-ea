# The model of the server

[`docs/mcp-server-sparx-ea.xml`](./mcp-server-sparx-ea.xml) is an
[ArchiMate](./glossary/archimate.md) model of this MCP server itself, exported in
[XMI](./glossary/xmi.md) 2.1. It is produced by the same service the MCP tools
call, so it doubles as a worked example of building a model from scratch.

## Import it

In Sparx EA: `File → Import → Package from XMI`, pick `docs/mcp-server-sparx-ea.xml`.
One root package **`Sparx EA MCP Server`** appears, with two sub-packages.

## What is inside

### `Возможности сервиса` — what the server does

| Element | Type | |
| --- | --- | --- |
| `mcp-server-sparx-ea` | ApplicationComponent | the server |
| 7 × capability | ApplicationFunction → ApplicationService | one per tool group |

The component is **assigned** to each function; each function **realizes** the
matching service. Every function's note lists its concrete `ea_*` tools and their
arguments. The groups: чтение модели, словарь типов, управление элементами,
управление связями, управление пакетами, размещение на диаграммах, жизненный цикл
модели.

### `План развития` — the MVP roadmap

Five `Plateau` elements, in order, linked by an `Association` named «затем»
(the ArchiMate validator only allows `Triggering` between behaviour elements):

1. **MVP** — *готово.* Read/write tools over an `Export Package to XMI` file.
2. **Публикация в GitHub workflow** — *готово.* Per-OS release binaries on a
   version bump; PR gate (unit tests + version-bump check).
3. **Создание Sparx XML** — *в работе.* `ea_new_model` / `ea_create_root_package` /
   `ea_set_root_name`; the ArchiMate3 profile is embedded in a from-scratch model.
4. **Отладка отчёта** — *запланировано.* `AssembleReport`; junit / coverage /
   mutation reports and README badges.
5. **Управление status** — *запланировано.* Read and change an element's `project`
   attributes (status / phase / version).

Each plateau has a `WorkPackage` (in `Работы`) that **realizes** it; where a
plateau delivers a capability, its work package also realizes that
`ApplicationService`.

## Regenerate

```
go run ./tools/modelgen              # writes docs/mcp-server-sparx-ea.xml
go run ./tools/modelgen -o other.xml
```
