Feature: Add an element to a diagram (method 6)

  Every scenario is an independent package inside tmp/diagram_object_add.xml.

  Background:
    Given the working model:
      | source | TestProject.xml        |
      | output | diagram_object_add.xml |

  Scenario: Add a new element to a diagram
    When I create a "ArchiMate.Requirement" named "PerfReq" in "Motivation_Package" with note ""
    And I add "Motivation_Package/PerfReq" to the diagram "Motivation_Package/Motivation_Diagram" at 100,100,200,170
    Then the placement succeeds
    And after reload the diagram "Motivation_Package/Motivation_Diagram" has 16 placed elements
    And after reload the diagram placed elements include:
      | name    | left | top | right | bottom |
      | PerfReq | 100  | 100 | 200   | 170    |

  Scenario: Adding the same element twice is refused
    When I add "Motivation_Package/Goal1" to the diagram "Motivation_Package/Motivation_Diagram" at 10,10,110,80
    Then the placement fails with "already on diagram"

  Scenario: An invalid rectangle is refused
    When I create a "ArchiMate.Goal" named "G2" in "Motivation_Package" with note ""
    And I add "Motivation_Package/G2" to the diagram "Motivation_Package/Motivation_Diagram" at 200,200,100,100
    Then the placement fails with "invalid rectangle"
