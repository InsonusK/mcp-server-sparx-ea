# ArchiMate

**ArchiMate** is an Open Group standard notation for enterprise architecture: a
fixed vocabulary of element types (Goal, Requirement, Business Process,
Application Component, Node, …) and relationship types (Realization, Serving,
Assignment, Composition, …) for describing how business, application and
technology layers fit together.

## Why it exists

UML is large and open-ended; two architects model the same enterprise very
differently. ArchiMate fixes a small, layered vocabulary so models are
comparable and the relationships between layers are explicit and rule-checked.

## How it works

- **Elements** are grouped into *aspects* — active structure (who/what acts),
  behaviour (what happens), passive structure (what is acted on), plus
  motivation, strategy, implementation.
- **Relationships** connect elements, and ArchiMate defines which relationship
  is allowed between which element types (the *derivation* / permission rules).
  For example, a behaviour element *realizes* a motivation element; you cannot
  *trigger* a passive object.
- Tools render elements with standard shapes and colours per layer.

ArchiMate 3.2 has roughly 60 element types and 11 relationship types plus a
Junction.

## How it is structured in this project

Sparx EA stores an ArchiMate element as a UML element with a stereotype:

| ArchiMate | EA on disk |
| --- | --- |
| `ArchiMate.Goal` | `uml:Class` + `stereotype = ArchiMate_Goal` |
| `ArchiMate.BusinessProcess` | `uml:Activity` + `stereotype = ArchiMate_BusinessProcess` (behaviour → Activity) |
| `ArchiMate.Realization` | `uml:Dependency` + `stereotype = ArchiMate_Realization` |
| `ArchiMate.Association` | `uml:Association` + `stereotype = ArchiMate_Association` |
| `ArchiMate.Influence` | `<edge xmi:type="uml:ControlFlow">` + `stereotype = ArchiMate_Influence` |

`internal/service/sparx/` hides that mapping: a caller says
`type = "ArchiMate.Goal"`, and the service:

1. checks the type is in the hard-coded ArchiMate 3.2 whitelist
   (`ea_archimate_types` returns the full list);
2. for a relationship, checks ArchiMate permits it between the two element
   types, and that no identical `(type, source, target)` already exists;
3. writes the right `uml:*` node, stereotype and `<ArchiMate3:*>` application.

## Example

```
ea_create_element  … type=ArchiMate.Goal  name="Reduce cost"
ea_create_relationship  … source=<a Requirement>  target=<that Goal>  type=ArchiMate.Realization
```

The second call is accepted because ArchiMate allows a Requirement to realize a
Goal; `type=ArchiMate.Triggering` between the same two would be refused
(`is not allowed: triggering connects two behaviour elements`).

## Related concepts

- [Sparx Enterprise Architect](sparx-enterprise-architect.md) — stores ArchiMate
  as UML-with-stereotypes.

## Sources

- The Open Group, *ArchiMate 3.2 Specification* — <https://pubs.opengroup.org/architecture/archimate3-doc/>
- The vocabulary and rules implemented in `internal/service/sparx/archimate.go`.
