Feature: Model navigator tree (method 1)
  As the MCP server
  I want the whole model as a tree of packages, diagrams and elements with IDs
  So that an agent can navigate the model like the Sparx EA project browser

  Background:
    Given the model file "TestProject.xml"

  Scenario: The root package chain
    When I read the model tree
    Then the node at path "" has children:
      | kind    | name  |
      | package | Model |
    And the node at path "Model" has children:
      | kind    | name               |
      | package | Motivation_Package |

  Scenario: A leaf package lists its diagram and every element with its ArchiMate type
    When I read the model tree
    Then the node at path "Model/Motivation_Package" has children:
      | kind    | name               | type                     |
      | diagram | Motivation_Diagram |                          |
      | element | Assessment1        | ArchiMate.Assessment     |
      | element | Capability1        | ArchiMate.Capability     |
      | element | Constraint1        | ArchiMate.Constraint     |
      | element | CourseOfAction1    | ArchiMate.CourseOfAction |
      | element | Driver1            | ArchiMate.Driver         |
      | element | Goal1              | ArchiMate.Goal           |
      | element | Location1          | ArchiMate.Location       |
      | element | Meaning1           | ArchiMate.Meaning        |
      | element | Outcome1           | ArchiMate.Outcome        |
      | element | Principle1         | ArchiMate.Principle      |
      | element | Requirement1       | ArchiMate.Requirement    |
      | element | Resource1          | ArchiMate.Resource       |
      | element | Stakeholder1       | ArchiMate.Stakeholder    |
      | element | Value1             | ArchiMate.Value          |
      | element | ValueStream1       | ArchiMate.ValueStream    |

  Scenario: Each node carries a stable EA id and GUID
    When I read the model tree
    Then the node at path "Model/Motivation_Package/Goal1" has id "EAID_F3C2209E_B146_4328_863E_622BABD2B43E"
    And the node at path "Model/Motivation_Package/Goal1" has guid "{F3C2209E-B146-4328-863E-622BABD2B43E}"

  Scenario: Deep package nesting is walked to any depth
    Given the model file "DeepModel.xml"
    When I read the model tree
    Then the node at path "DeepModel/Package1/Package2/Package3" has children:
      | kind    | name            | type               |
      | diagram | Package3Diagram |                    |
      | element | Goal1           | ArchiMate.Goal     |
      | element | Resource1       | ArchiMate.Resource |

  Scenario: Cyrillic package and element names round-trip in the tree
    Given the model file "CyrillicProject.xml"
    When I read the model tree
    Then the node at path "Модель/Пакет" has children:
      | kind    | name      | type           |
      | diagram | Диаграмма |                |
      | element | Цель      | ArchiMate.Goal |
