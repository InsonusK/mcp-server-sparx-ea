Feature: Delete an element (method 4)

  Each scenario declares its fixture and its own tmp file.

  Scenario: Deleting an element also removes every relationship attached to it
    Given the working model:
      | source | TestProject.xml    |
      | output | element_delete.xml |
      | root   | element delete     |
    When I delete the element "element delete/Motivation_Package/Goal1"
    Then the delete succeeds
    And after reload the element "element delete/Motivation_Package/Goal1" cannot be found
    And after reload the element "element delete/Motivation_Package/Assessment1" has no relation to "Goal1"
    And after reload the element "element delete/Motivation_Package/Requirement1" has no relation to "Goal1"

  Scenario: Deleting a missing element is an error
    Given the working model:
      | source | TestProject.xml        |
      | output | element_delete_bad.xml |
      | root   | element delete bad     |
    When I delete the element "element delete bad/Motivation_Package/Ghost"
    Then the delete fails with "no element for"
