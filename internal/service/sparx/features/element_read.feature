Feature: Read element information (method 2)
  As the MCP server
  I want an element's ArchiMate type, note and relationships by id or by path
  So that an agent can inspect one element without walking the whole model

  Background:
    Given the model file "TestProject.xml"

  Scenario: Read an element by path
    When I read the element "Model/Motivation_Package/Goal1"
    Then the element is:
      | field         | value                                    |
      | id            | EAID_F3C2209E_B146_4328_863E_622BABD2B43E |
      | guid          | {F3C2209E-B146-4328-863E-622BABD2B43E}    |
      | name          | Goal1                                    |
      | type          | ArchiMate.Goal                           |
      | path          | Model/Motivation_Package/Goal1           |
      | documentation | Goal description                         |

  Scenario: The same element is reachable by GUID
    When I read the element "{F3C2209E-B146-4328-863E-622BABD2B43E}"
    Then the element field "name" is "Goal1"
    And the element field "type" is "ArchiMate.Goal"

  Scenario: An element reports every relationship with direction and the other end
    When I read the element "Model/Motivation_Package/Goal1"
    Then the element relations are exactly:
      | type                  | direction | otherName    | otherType             |
      | ArchiMate.Influence   | incoming  | Assessment1  | ArchiMate.Assessment  |
      | ArchiMate.Realization | incoming  | Constraint1  | ArchiMate.Constraint  |
      | ArchiMate.Realization | incoming  | Outcome1     | ArchiMate.Outcome     |
      | ArchiMate.Realization | incoming  | Principle1   | ArchiMate.Principle   |
      | ArchiMate.Realization | incoming  | Requirement1 | ArchiMate.Requirement |

  Scenario: A named relationship keeps its name and outgoing direction from its source
    When I read the element "Model/Motivation_Package/Driver1"
    Then the element relations include:
      | type                | name                 | direction | otherName   |
      | ArchiMate.Influence | Driver to Assessment | outgoing  | Assessment1 |

  Scenario: Read a Cyrillic element
    Given the model file "CyrillicProject.xml"
    When I read the element "Модель/Пакет/Цель"
    Then the element is:
      | field         | value          |
      | name          | Цель           |
      | type          | ArchiMate.Goal |
      | documentation | Описание Цели   |

  Scenario: An element whose stereotype is not ArchiMate is reported as unknown, not mislabelled
    Given the model file "PlainUML.xml"
    When I read the element "Plain/PlainClass"
    Then the element is:
      | field | value         |
      | type  | unknown:Class |
    And the element relations are exactly:
      | type               | direction | otherName      | otherType         |
      | unknown:Dependency | outgoing  | PlainComponent | unknown:Component |

  Scenario: Reading a missing element is an error
    When I read the element "Model/Motivation_Package/DoesNotExist"
    Then the read fails with "no element for"

  Scenario: A relationship the ArchiMate rules forbid is flagged on the element
    Given the model file "rule_violations.xml"
    When I read the element "Rule violations/Elements/Goal1"
    Then the element relations include:
      | type                 | direction | otherName | verdict |
      | ArchiMate.Assignment | outgoing  | Node1     | deny    |
      | ArchiMate.Realization | incoming  | Requirement1 |      |
