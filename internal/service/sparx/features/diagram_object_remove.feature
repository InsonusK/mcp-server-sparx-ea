Feature: Remove an element from a diagram (method 6)

  Each scenario declares its fixture and its own tmp file.

  Scenario: Removing an element from a diagram leaves the element in the model
    Given the working model:
      | source | TestProject.xml           |
      | output | diagram_object_remove.xml |
      | root   | diagram object remove     |
    When I remove "diagram object remove/Motivation_Package/Goal1" from the diagram "diagram object remove/Motivation_Package/Motivation_Diagram"
    Then the placement succeeds
    And after reload the diagram "diagram object remove/Motivation_Package/Motivation_Diagram" has 14 placed elements
    And after reload the element "diagram object remove/Motivation_Package/Goal1" is:
      | field | value          |
      | type  | ArchiMate.Goal |

  Scenario: Removing an element that is not on the diagram is refused
    Given the working model:
      | source | TestProject.xml                  |
      | output | diagram_object_remove_absent.xml |
      | root   | diagram object remove absent     |
    When I create a "ArchiMate.Goal" named "NotShown" in "diagram object remove absent/Motivation_Package" with note ""
    And I remove "diagram object remove absent/Motivation_Package/NotShown" from the diagram "diagram object remove absent/Motivation_Package/Motivation_Diagram"
    Then the placement fails with "not on diagram"
