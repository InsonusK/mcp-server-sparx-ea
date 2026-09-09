Feature: Create elements (method 4)
  As the MCP server
  I want validated element creation
  So that an agent can add elements without corrupting the model

  All successful creates accumulate into tmp/element_create.xml.

  Background:
    Given the working model:
      | source | TestProject.xml    |
      | output | element_create.xml |
      | root   | element create     |

  Scenario: Create an ArchiMate element in a package
    When I create a "ArchiMate.Requirement" named "AvailabilityReq" in "element create/Motivation_Package" with note "must be 99.9% available"
    Then the create succeeds
    And after reload the element "element create/Motivation_Package/AvailabilityReq" is:
      | field         | value                  |
      | type          | ArchiMate.Requirement  |
      | documentation | must be 99.9% available |

  Scenario: A behaviour element keeps the ArchiMate type even though EA's base type is uml:Activity
    When I create a "ArchiMate.BusinessProcess" named "Onboarding" in "element create/Motivation_Package" with note ""
    Then the create succeeds
    And after reload the element field "type" is "ArchiMate.BusinessProcess"

  Scenario Outline: Creating an element outside the ArchiMate vocabulary is refused
    When I create a "<type>" named "X" in "element create/Motivation_Package" with note ""
    Then the create fails with "not a known ArchiMate element type"

    Examples:
      | type                  |
      | ArchiMate.Frobnicator |
      | uml:Class             |
      | ArchiMate.Junction    |

  Scenario: A duplicate name in the same package is refused
    When I create a "ArchiMate.Goal" named "Goal1" in "element create/Motivation_Package" with note ""
    Then the create fails with "already has an element named"
