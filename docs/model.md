# The model of the server

[`docs/mcp-server-sparx-ea.arch.xml`](./mcp-server-sparx-ea.arch.xml) is an
[ArchiMate](./glossary/archimate.md) model of this MCP server itself, exported in
[XMI](./glossary/xmi.md) 2.1. It is produced by the same service the MCP tools
call (`ea_new_model` / `ea_create_*` / `ea_create_diagram` / `ea_place_on_diagram`),
so it doubles as a worked example of building a model — diagrams included — from
scratch.

## Import it

In Sparx EA: `File → Import → Package from XMI`, pick
`docs/mcp-server-sparx-ea.arch.xml`. One root package **`Sparx EA MCP Server`**
appears, with three sub-packages and two diagrams.

## What is inside

### `Возможности сервиса` — what the server does

| Element | Type | |
| --- | --- | --- |
| `mcp-server-sparx-ea` | ApplicationComponent | the server |
| 9 × capability | ApplicationFunction → ApplicationService | one per tool group |

The component is **assigned** to each function; each function **realizes** the
matching service. Every function's note lists its concrete `ea_*` tools and their
arguments. The groups: чтение модели, словарь типов, управление элементами,
управление связями, управление пакетами, диаграммы, жизненный цикл модели, and
the planned «отправка bug report» и «запрос новых типов ArchiMate».

The **`Обзор возможностей`** diagram (Application viewpoint) shows the component,
its functions and the services they realize.

### `План развития` — the roadmap

Seven `Plateau` elements, in order, linked by an `Association` named «затем»
(the ArchiMate validator only allows `Triggering` between behaviour elements):

1. **MVP** — *готово.* Read/write tools over an `Export Package to XMI` file.
2. **Публикация в GitHub** — *готово.* Per-OS release binaries on a version bump;
   one-line installers; PR gate (unit tests + version-bump check).
3. **Доп функции** — *в работе.* `ea_new_model` / `ea_create_root_package` /
   `ea_set_root_name` (ArchiMate3 profile embedded); `ea_copy_package`;
   `ea_create_diagram` + placing elements on diagrams.
4. **Улучшения отчётов о тестировании** — *запланировано.* junit / coverage /
   mutation reports and README badges on GitHub Pages; `AssembleReport`.
5. **Нумерация всех ошибок** — *запланировано.* Stable `EA-NNNN` error codes across
   MCP messages, logs and docs; a known-issues registry.
6. **Отправка bug report** — *запланировано.* A tool the agent uses to file a bug
   report with the maintainer (error code, call context, input file).
7. **Запрос добавления типов** — *запланировано.* A request to extend the type
   vocabulary when the server meets an unknown stereotype on read, or the agent
   needs a type the vocabulary lacks.

Each plateau has a `WorkPackage` (in `Работы`) that **realizes** it; where a
plateau delivers a capability, its work package also realizes that
`ApplicationService`. The `Требования` package holds one `Requirement`,
**`Сквозная нумерация ошибок`**, realized by step 5's work package.

The **`Дорожная карта`** diagram (Implementation & Migration viewpoint) lays the
plateaus in a row with their work packages beneath them.

## Regenerate

```
go run ./tools/modelgen              # writes docs/mcp-server-sparx-ea.arch.xml
go run ./tools/modelgen -o other.xml
```
