Feature: The service speaks ArchiMate, not (uml:Class + stereotype)
  Architectural concern: an ArchiMate element in EA is stored as a base UML type
  plus a stereotype ("uml:Class" + "ArchiMate_Goal"). The service must hide that
  split — callers see one type, "ArchiMate.Goal" — and must not silently
  mislabel a type it does not recognise.

  Background:
    Given the model file "TestProject.xml"

  Scenario Outline: EA stereotype becomes a single ArchiMate type
    When I read the element "Model/Motivation_Package/<element>"
    Then the element field "type" is "<type>"

    Examples:
      | element      | type                  |
      | Goal1        | ArchiMate.Goal        |
      | ValueStream1 | ArchiMate.ValueStream |
      | Stakeholder1 | ArchiMate.Stakeholder |

  Scenario Outline: EA connector stereotype becomes a single ArchiMate relationship type
    When I read the element "Model/Motivation_Package/<element>"
    Then the element relations include:
      | type   |
      | <type> |

    Examples:
      | element    | type                  |
      | Goal1      | ArchiMate.Influence   |
      | Goal1      | ArchiMate.Realization |
      | Driver1    | ArchiMate.Association  |

  Scenario: A whole ArchiMate 3.2 element type is in the vocabulary
    Then "ArchiMate.BusinessProcess" is a known element type
    And "ArchiMate.ApplicationComponent" is a known element type
    And "ArchiMate.Node" is a known element type

  Scenario: A type outside ArchiMate is not a known element type
    Then "ArchiMate.Frobnicator" is not a known element type
    And "uml:Class" is not a known element type

  Scenario: An element with no ArchiMate stereotype is reported as unknown, not mislabelled
    Given the model file "PlainUML.xml"
    When I read the element "Plain/PlainClass"
    Then the element is:
      | field         | value                            |
      | type          | unknown:Class                    |
      | documentation | a plain class with no stereotype |

  Scenario: A relationship with no ArchiMate stereotype is reported as unknown
    Given the model file "PlainUML.xml"
    When I read the element "Plain/PlainClass"
    Then the element relations are exactly:
      | type              | direction | otherName      | otherType        |
      | unknown:Dependency | outgoing  | PlainComponent | unknown:Component |
