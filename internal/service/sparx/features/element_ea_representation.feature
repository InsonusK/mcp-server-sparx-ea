Feature: An element's EA representation matches its ArchiMate type (method 4)
  Architectural concern: Sparx EA only renders an element with its real
  ArchiMate shape (interface, component, business process, plain class, …)
  when the stereotype is applied to the UML base type EA's own ArchiMate3
  profile expects for it — an "ArchiMate_TechnologyInterface" stereotype only
  registers on a uml:Interface base (base_Interface). Applied to a uml:Class
  it renders in EA as an anonymous, unstyled class instead of a Technology
  Interface. tmp/examples/sandbox.xml is a real EA export of one of every
  ArchiMate type, used to derive the expected base type below.

  Scenario Outline: An ArchiMate element gets EA's correct UML base type
    Given the working model:
      | source | TestProject.xml     |
      | output | element_ea_repr.xml |
      | root   | ea repr <type>      |
    When I create a "ArchiMate.<type>" named "<type>" in "ea repr <type>/Motivation_Package" with note ""
    Then the create succeeds
    And after reload the element "ea repr <type>/Motivation_Package/<type>" has EA type "<umlType>"

    Examples:
      | type                 | umlType       |
      | TechnologyInterface  | uml:Interface |
      | ApplicationInterface | uml:Interface |
      | BusinessInterface    | uml:Interface |
      | ApplicationComponent | uml:Component |
      | ValueStream          | uml:Activity  |
      | ImplementationEvent  | uml:Activity  |
      | BusinessProcess      | uml:Activity  |
      | Goal                 | uml:Class     |
      | Node                 | uml:Class     |

  Scenario: Every ArchiMate element type creates as a correctly-shaped EA element
    One of every ArchiMate 3.2 element type, all in a single package. Saved to
    tmp/scenario/element_ea_repr_all.xml (part of tmp/report.xml) — import it
    into Sparx EA to see every element rendered with its real ArchiMate shape
    instead of as an anonymous class.

    Given the working model:
      | source | TestProject.xml         |
      | output | element_ea_repr_all.xml |
      | root   | ArchimateElements       |
    When I create one element of every ArchiMate type in "ArchimateElements/Motivation_Package"
    Then the create succeeds
    And after reload every element in "ArchimateElements/Motivation_Package" has EA's correct UML base type
