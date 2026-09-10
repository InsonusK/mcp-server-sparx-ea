# Example models — the relationship rules

Three tiny EA "Import Package from XMI" models, one per verdict of the
[ArchiMate relationship rules](../archimate-relationship-matrix.md)
([`internal/service/sparx/archimate_relationships.csv`](../../internal/service/sparx/archimate_relationships.csv)):

| File | Contains |
| --- | --- |
| `relationships-allow.xml` | relationships the rules **allow** |
| `relationships-warn.xml` | relationships the rules only **warn** about |
| `relationships-deny.xml` | relationships the rules **deny** |

Regenerate with `go run ./tools/relexamples`.

## Try it

```
ea_model_tree      file=docs/examples/relationships-deny.xml
    → root node carries "notices": ["5 relationship(s) violate the ArchiMate rules …"]

ea_element         file=docs/examples/relationships-warn.xml  ref="Relationships — warn/Elements/Grouping1"
    → the Aggregation relation carries "verdict": "warn" and a "warning" string

ea_create_relationship  file=docs/examples/relationships-allow.xml  output=/tmp/e.xml \
    source="Relationships — allow/Elements/Goal1" target="Relationships — allow/Elements/Node1" \
    type=ArchiMate.Assignment
    → refused: "not allowed by the ArchiMate relationship rules"  (Goal → Node Assignment is deny)
```
