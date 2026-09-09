Feature: Add, move and remove elements on a diagram (method 6)

  Background:
    Given a working copy of "TestProject.xml"

  Scenario: Add an element to a diagram
    When I create a "ArchiMate.Requirement" named "PerfReq" in "Model/Motivation_Package" with note ""
    And I add "Model/Motivation_Package/PerfReq" to the diagram "Model/Motivation_Package/Motivation_Diagram" at 100,100,200,170
    Then the placement succeeds
    And after reload the diagram "Model/Motivation_Package/Motivation_Diagram" has 16 placed elements
    And after reload the diagram placed elements include:
      | name    | left | top | right | bottom |
      | PerfReq | 100  | 100 | 200   | 170    |

  Scenario: Adding the same element twice is refused
    When I add "Model/Motivation_Package/Goal1" to the diagram "Model/Motivation_Package/Motivation_Diagram" at 10,10,110,80
    Then the placement fails with "already on diagram"

  Scenario: An invalid rectangle is refused
    When I create a "ArchiMate.Goal" named "G2" in "Model/Motivation_Package" with note ""
    And I add "Model/Motivation_Package/G2" to the diagram "Model/Motivation_Package/Motivation_Diagram" at 200,200,100,100
    Then the placement fails with "invalid rectangle"

  Scenario: Move an element on a diagram
    When I move "Model/Motivation_Package/Goal1" on the diagram "Model/Motivation_Package/Motivation_Diagram" to 300,300,400,370
    Then the placement succeeds
    And after reload the diagram placed elements include:
      | name  | left | top | right | bottom |
      | Goal1 | 300  | 300 | 400   | 370    |

  Scenario: Remove an element from a diagram (the element stays in the model)
    When I remove "Model/Motivation_Package/Goal1" from the diagram "Model/Motivation_Package/Motivation_Diagram"
    Then the placement succeeds
    And after reload the diagram "Model/Motivation_Package/Motivation_Diagram" has 14 placed elements
    And after reload the element "Model/Motivation_Package/Goal1" is:
      | field | value          |
      | type  | ArchiMate.Goal |
