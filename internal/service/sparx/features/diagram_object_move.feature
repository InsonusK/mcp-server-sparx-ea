Feature: Move an element on a diagram (method 6)

  Every scenario is an independent package inside tmp/diagram_object_move.xml.

  Background:
    Given the working model:
      | source | TestProject.xml         |
      | output | diagram_object_move.xml |

  Scenario: Move an element to a new rectangle
    When I move "Motivation_Package/Goal1" on the diagram "Motivation_Package/Motivation_Diagram" to 300,300,400,370
    Then the placement succeeds
    And after reload the diagram "Motivation_Package/Motivation_Diagram" has 15 placed elements
    And after reload the diagram placed elements include:
      | name  | left | top | right | bottom |
      | Goal1 | 300  | 300 | 400   | 370    |

  Scenario: Moving an element that is not on the diagram is refused
    When I create a "ArchiMate.Goal" named "Offdiagram" in "Motivation_Package" with note ""
    And I move "Motivation_Package/Offdiagram" on the diagram "Motivation_Package/Motivation_Diagram" to 10,10,50,50
    Then the placement fails with "not on diagram"
