Feature: Remove an element from a diagram (method 6)

  Every scenario is an independent package inside tmp/diagram_object_remove.xml.

  Background:
    Given the working model:
      | source | TestProject.xml           |
      | output | diagram_object_remove.xml |

  Scenario: Removing an element from a diagram leaves the element in the model
    When I remove "Motivation_Package/Goal1" from the diagram "Motivation_Package/Motivation_Diagram"
    Then the placement succeeds
    And after reload the diagram "Motivation_Package/Motivation_Diagram" has 14 placed elements
    And after reload the element "Motivation_Package/Goal1" is:
      | field | value          |
      | type  | ArchiMate.Goal |

  Scenario: Removing an element that is not on the diagram is refused
    When I create a "ArchiMate.Goal" named "NotShown" in "Motivation_Package" with note ""
    And I remove "Motivation_Package/NotShown" from the diagram "Motivation_Package/Motivation_Diagram"
    Then the placement fails with "not on diagram"
