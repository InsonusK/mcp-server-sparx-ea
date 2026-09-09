Feature: Create elements (method 4)
  As the MCP server
  I want validated element creation
  So that an agent can add elements without corrupting the model

  Each scenario is an independent package inside tmp/element_create.xml (import
  it once; every scenario appears as its own numbered root package).

  Background:
    Given the working model:
      | source | TestProject.xml    |
      | output | element_create.xml |

  Scenario: Create an ArchiMate element in a package
    When I create a "ArchiMate.Requirement" named "AvailabilityReq" in "Motivation_Package" with note "must be 99.9% available"
    Then the create succeeds
    And after reload the element "Motivation_Package/AvailabilityReq" is:
      | field         | value                  |
      | type          | ArchiMate.Requirement  |
      | documentation | must be 99.9% available |

  Scenario: A behaviour element keeps the ArchiMate type even though EA's base type is uml:Activity
    When I create a "ArchiMate.BusinessProcess" named "Onboarding" in "Motivation_Package" with note ""
    Then the create succeeds
    And after reload the element field "type" is "ArchiMate.BusinessProcess"

  Scenario Outline: Creating an element outside the ArchiMate vocabulary is refused
    When I create a "<type>" named "X" in "Motivation_Package" with note ""
    Then the create fails with "not a known ArchiMate element type"

    Examples:
      | type                  |
      | ArchiMate.Frobnicator |
      | uml:Class             |
      | ArchiMate.Junction    |

  Scenario: A duplicate name in the same package is refused
    When I create a "ArchiMate.Goal" named "Goal1" in "Motivation_Package" with note ""
    Then the create fails with "already has an element named"
