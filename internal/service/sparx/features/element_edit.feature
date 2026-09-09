Feature: Edit an element's basic properties (method 4)

  Every scenario is an independent package inside tmp/element_edit.xml.

  Background:
    Given the working model:
      | source | TestProject.xml  |
      | output | element_edit.xml |

  Scenario: Rename an element and change its note
    When I rename the element "Motivation_Package/Value1" to "CustomerValue"
    And I set the note of "Motivation_Package/CustomerValue" to "value delivered to the customer"
    Then the rename succeeds
    And after reload the element "Motivation_Package/CustomerValue" is:
      | field         | value                           |
      | name          | CustomerValue                   |
      | type          | ArchiMate.Value                 |
      | documentation | value delivered to the customer |

  Scenario: Renaming to a blank name is refused
    When I rename the element "Motivation_Package/Value1" to "   "
    Then the rename fails with "new name is required"
