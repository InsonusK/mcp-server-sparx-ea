Feature: Read package contents
  As the MCP server
  I want a package's name, path, parent and immediate children by id or path
  So that an agent can inspect one package without walking the whole tree

  Background:
    Given the model file "TestProject.xml"

  Scenario: Read a leaf package by path
    When I read the package "Model/Motivation_Package"
    Then the package is:
      | field  | value                    |
      | name   | Motivation_Package       |
      | path   | Model/Motivation_Package |
      | parent | Model                    |
    And the package packages are ""
    And the package diagrams are "Motivation_Diagram"
    And the package elements are "Stakeholder1, Driver1, Assessment1, Goal1, Outcome1, Principle1, Requirement1, Constraint1, Meaning1, Value1, Resource1, Capability1, CourseOfAction1, Location1, ValueStream1"

  Scenario: Read the EA root package
    When I read the package "Model"
    Then the package is:
      | field  | value |
      | name   | Model |
      | path   | Model |
      | parent |       |
    And the package packages are "Motivation_Package"

  Scenario: The same package is reachable by GUID
    When I read the package "{91AD8E48-3AE7-49f7-AAAB-72CA99521F53}"
    Then the package is:
      | field | value              |
      | name  | Motivation_Package |

  Scenario: Deep package nesting reports the right parent path
    Given the model file "DeepModel.xml"
    When I read the package "DeepModel/Package1/Package2/Package3"
    Then the package is:
      | field  | value                            |
      | path   | DeepModel/Package1/Package2/Package3 |
      | parent | DeepModel/Package1/Package2       |
    And the package elements are "Goal1, Resource1"
    And the package diagrams are "Package3Diagram"

  Scenario: Reading a missing package is an error
    When I read the package "Model/Nope"
    Then the read fails with "no package for"
