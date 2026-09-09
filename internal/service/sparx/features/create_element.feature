Feature: Create, rename, note and delete elements (method 4)
  As the MCP server
  I want validated element edits written to a working copy
  So that an agent can change the model without corrupting it

  Background:
    Given a working copy of "TestProject.xml"

  Scenario: Create an ArchiMate element in a package
    When I create a "ArchiMate.Requirement" named "AvailabilityReq" in "Model/Motivation_Package" with note "must be 99.9% available"
    Then the create succeeds
    And after reload the element "Model/Motivation_Package/AvailabilityReq" is:
      | field         | value                     |
      | type          | ArchiMate.Requirement     |
      | documentation | must be 99.9% available    |

  Scenario: Creating a behaviour element uses the right base type but the same ArchiMate type
    When I create a "ArchiMate.BusinessProcess" named "Onboarding" in "Model/Motivation_Package" with note ""
    Then the create succeeds
    And after reload the element field "type" is "ArchiMate.BusinessProcess"

  Scenario Outline: Creating an element outside the ArchiMate vocabulary is refused
    When I create a "<type>" named "X" in "Model/Motivation_Package" with note ""
    Then the create fails with "<message>"

    Examples:
      | type                  | message                                  |
      | ArchiMate.Frobnicator | not a known ArchiMate element type        |
      | uml:Class             | not a known ArchiMate element type        |
      | ArchiMate.Junction    | not a known ArchiMate element type        |

  Scenario: A duplicate name in the same package is refused
    When I create a "ArchiMate.Goal" named "Goal1" in "Model/Motivation_Package" with note ""
    Then the create fails with "already has an element named"

  Scenario: Rename and re-note an element
    When I rename "Model/Motivation_Package/Value1" to "CustomerValue"
    And I set the note of "Model/Motivation_Package/CustomerValue" to "value to the customer"
    Then after reload the element "Model/Motivation_Package/CustomerValue" is:
      | field         | value                 |
      | name          | CustomerValue         |
      | documentation | value to the customer |

  Scenario: Delete an element and its relationships
    When I delete the element "Model/Motivation_Package/Goal1"
    Then the delete succeeds
    And after reload the element "Model/Motivation_Package/Goal1" cannot be found
    And after reload the element "Model/Motivation_Package/Assessment1" has no relation to "Goal1"
