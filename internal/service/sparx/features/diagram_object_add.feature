Feature: Add an element to a diagram (method 6)

  Background:
    Given the working model:
      | source | TestProject.xml        |
      | output | diagram_object_add.xml |
      | root   | diagram object add     |

  Scenario: Add a new element to a diagram
    When I create a "ArchiMate.Requirement" named "PerfReq" in "diagram object add/Motivation_Package" with note ""
    And I add "diagram object add/Motivation_Package/PerfReq" to the diagram "diagram object add/Motivation_Package/Motivation_Diagram" at 100,100,200,170
    Then the placement succeeds
    And after reload the diagram "diagram object add/Motivation_Package/Motivation_Diagram" has 16 placed elements
    And after reload the diagram placed elements include:
      | name    | left | top | right | bottom |
      | PerfReq | 100  | 100 | 200   | 170    |

  Scenario: Adding the same element twice is refused
    When I add "diagram object add/Motivation_Package/Goal1" to the diagram "diagram object add/Motivation_Package/Motivation_Diagram" at 10,10,110,80
    Then the placement fails with "already on diagram"

  Scenario: An invalid rectangle is refused
    When I create a "ArchiMate.Goal" named "G2" in "diagram object add/Motivation_Package" with note ""
    And I add "diagram object add/Motivation_Package/G2" to the diagram "diagram object add/Motivation_Package/Motivation_Diagram" at 200,200,100,100
    Then the placement fails with "invalid rectangle"
