Feature: Edit an element's basic properties (method 4)

  Background:
    Given the working model:
      | source | TestProject.xml  |
      | output | element_edit.xml |
      | root   | element edit     |

  Scenario: Rename an element and change its note
    When I rename the element "element edit/Motivation_Package/Value1" to "CustomerValue"
    And I set the note of "element edit/Motivation_Package/CustomerValue" to "value delivered to the customer"
    Then the rename succeeds
    And after reload the element "element edit/Motivation_Package/CustomerValue" is:
      | field         | value                          |
      | name          | CustomerValue                  |
      | type          | ArchiMate.Value                |
      | documentation | value delivered to the customer |

  Scenario: Renaming to a blank name is refused
    When I rename the element "element edit/Motivation_Package/Value1" to "   "
    Then the rename fails with "new name is required"
