Feature: Edit an element's basic properties (method 4)

  Each scenario declares its fixture and its own tmp file.

  Scenario: Rename an element and change its note
    Given the working model:
      | source | TestProject.xml  |
      | output | element_edit.xml |
      | root   | element edit     |
    When I rename the element "element edit/Motivation_Package/Value1" to "CustomerValue"
    And I set the note of "element edit/Motivation_Package/CustomerValue" to "value delivered to the customer"
    Then the rename succeeds
    And after reload the element "element edit/Motivation_Package/CustomerValue" is:
      | field         | value                           |
      | name          | CustomerValue                   |
      | type          | ArchiMate.Value                 |
      | documentation | value delivered to the customer |

  Scenario: Renaming to a blank name is refused
    Given the working model:
      | source | TestProject.xml      |
      | output | element_edit_bad.xml |
      | root   | element edit bad     |
    When I rename the element "element edit bad/Motivation_Package/Value1" to "   "
    Then the rename fails with "new name is required"
