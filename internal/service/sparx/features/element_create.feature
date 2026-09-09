Feature: Create elements (method 4)
  As the MCP server
  I want validated element creation
  So that an agent can add elements without corrupting the model

  Each scenario declares its fixture and its own tmp file.

  Scenario: Create an ArchiMate element in a package
    Given the working model:
      | source | TestProject.xml    |
      | output | element_create.xml |
      | root   | element create     |
    When I create a "ArchiMate.Requirement" named "AvailabilityReq" in "element create/Motivation_Package" with note "must be 99.9% available"
    Then the create succeeds
    And after reload the element "element create/Motivation_Package/AvailabilityReq" is:
      | field         | value                   |
      | type          | ArchiMate.Requirement   |
      | documentation | must be 99.9% available |

  Scenario: A behaviour element keeps the ArchiMate type even though EA's base type is uml:Activity
    Given the working model:
      | source | TestProject.xml              |
      | output | element_create_behaviour.xml |
      | root   | element create behaviour     |
    When I create a "ArchiMate.BusinessProcess" named "Onboarding" in "element create behaviour/Motivation_Package" with note ""
    Then the create succeeds
    And after reload the element field "type" is "ArchiMate.BusinessProcess"

  Scenario Outline: Creating an element outside the ArchiMate vocabulary is refused
    Given the working model:
      | source | TestProject.xml        |
      | output | element_create_bad.xml |
      | root   | element create bad     |
    When I create a "<type>" named "X" in "element create bad/Motivation_Package" with note ""
    Then the create fails with "not a known ArchiMate element type"

    Examples:
      | type                  |
      | ArchiMate.Frobnicator |
      | uml:Class             |
      | ArchiMate.Junction    |

  Scenario: A duplicate name in the same package is refused
    Given the working model:
      | source | TestProject.xml        |
      | output | element_create_dup.xml |
      | root   | element create dup     |
    When I create a "ArchiMate.Goal" named "Goal1" in "element create dup/Motivation_Package" with note ""
    Then the create fails with "already has an element named"
