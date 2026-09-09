# The editing workflow

Sparx EA 15.2 stores a project in an Access database that cannot be written
safely from outside EA. So the ArchiMate editing tools do **not** touch the
`.eapx` file. Instead they work on an [XMI](glossary/xmi.md) export, and you
re-import the result. The loop is:

```
                 ┌─────────────────────────────────────────────┐
                 │  1. In EA: File → Export → Package to XMI    │
                 │     (XMI 2.1) → model.xml                    │
                 └─────────────────────────────────────────────┘
                                     │
                                     ▼
   ┌──────────────────────────────────────────────────────────────┐
   │  2. Agent calls editing tools:                                │
   │       ea_create_element  file=model.xml  output=edited.xml …  │
   │       ea_create_relationship  file=edited.xml  output=…       │
   │     Each tool opens `file`, applies one change, writes        │
   │     `output`. Chain them by feeding each output to the next   │
   │     `file`.                                                   │
   └──────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
   ┌──────────────────────────────────────────────────────────────┐
   │  3. In EA: File → Import → Package from XMI → edited.xml       │
   │     EA diffs by GUID, shows what will change, asks to confirm │
   └──────────────────────────────────────────────────────────────┘
```

## Why a separate output file

`output` must differ from `file`, and neither is the `.eapx`. Keeping the
original export untouched means:

- you can inspect the diff (`edited.xml` vs `model.xml`) before importing;
- a tool that fails half-way leaves the input intact;
- if the import misbehaves, the exact file that caused it is on disk.

## Identity: importing next to the original

An XMI import matches existing model content **by GUID**. If you export a
package, edit the copy, and re-import it, EA updates the original package in
place. If you want the edited copy to land in EA as a *separate* package
alongside the original, the model needs fresh GUIDs — the ArchiMate service
does this when it renames the root package with a fresh identity (used by the
test suite; not yet exposed as a tool).

## report.xml

The test suite produces `internal/service/sparx/test/tmp/report.xml`: every
editing scenario merged into one importable file, each scenario a sub-package
of a single root, all with fresh GUIDs. It is a worked example of what every
tool produces — import it into a throwaway EA project to see the effect of the
whole tool set at once. Import **either** `report.xml` **or** the individual
`tmp/scenario/*.xml` files into one project, never both (they share GUIDs).

## Limits of the current tools

- Element basics only: name, ArchiMate type, note, parent package. No tagged
  values, no attributes/operations.
- Relationship basics only: type, source, target, name, note.
- `(type, source, target)` is a unique key — a duplicate relationship is
  refused.
- Diagrams: place / move / remove an existing element. No diagram creation, no
  connector routing, no styling.
- Only ArchiMate 3.2 element and relationship types are accepted; UML and BPMN
  are out of scope for now.
