Feature: Create a diagram (method 6)

  Each scenario declares its fixture and its own tmp file, so the file always
  shows exactly that scenario's result.

  Scenario: Create an ArchiMate diagram and place an element on it
    Given the working model:
      | source | TestProject.xml     |
      | output | diagram_create.xml  |
      | root   | diagram create      |
    When I create a diagram "Capabilities" with layer "Application" in "diagram create/Motivation_Package"
    Then the create succeeds
    And the diagram field "name" is "Capabilities"
    And the diagram field "path" is "diagram create/Motivation_Package/Capabilities"
    And the diagram field "diagramType" is "Logical"
    When I add "diagram create/Motivation_Package/Goal1" to the diagram "diagram create/Motivation_Package/Capabilities" at 60,60,220,140
    Then the placement succeeds
    And after reload the package "diagram create/Motivation_Package" diagrams are "Motivation_Diagram, Capabilities"
    And after reload the diagram "diagram create/Motivation_Package/Capabilities" has 1 placed elements
    And after reload the diagram "diagram create/Motivation_Package/Capabilities" placed elements include:
      | name  | left | top | right | bottom |
      | Goal1 | 60   | 60  | 220   | 140    |

  Scenario: Placing two connected elements shows the connector between them
    Given the working model:
      | source | TestProject.xml           |
      | output | diagram_create_links.xml  |
      | root   | diagram create links      |
    When I create a diagram "Realisation" with layer "Motivation" in "diagram create links/Motivation_Package"
    And I add "diagram create links/Motivation_Package/Goal1" to the diagram "diagram create links/Motivation_Package/Realisation" at 60,60,220,140
    And I add "diagram create links/Motivation_Package/Requirement1" to the diagram "diagram create links/Motivation_Package/Realisation" at 60,240,220,320
    Then the placement succeeds
    And after reload the diagram "diagram create links/Motivation_Package/Realisation" has 2 placed elements
    And after reload the diagram "diagram create links/Motivation_Package/Realisation" has 1 links
    When I remove "diagram create links/Motivation_Package/Requirement1" from the diagram "diagram create links/Motivation_Package/Realisation"
    Then after reload the diagram "diagram create links/Motivation_Package/Realisation" has 0 links

  Scenario: Create a plain diagram without a layer
    Given the working model:
      | source | TestProject.xml          |
      | output | diagram_create_plain.xml |
      | root   | diagram create plain     |
    When I create a diagram "Notes" with layer "" in "diagram create plain/Motivation_Package"
    Then the create succeeds
    And after reload the package "diagram create plain/Motivation_Package" diagrams are "Motivation_Diagram, Notes"

  Scenario: An unknown layer is refused
    Given the working model:
      | source | TestProject.xml        |
      | output | diagram_create_bad.xml |
      | root   | diagram create bad     |
    When I create a diagram "Bad" with layer "Nonsense" in "diagram create bad/Motivation_Package"
    Then the create fails with "not a known ArchiMate diagram layer"

  Scenario: A blank diagram name is refused
    Given the working model:
      | source | TestProject.xml          |
      | output | diagram_create_blank.xml |
      | root   | diagram create blank     |
    When I create a diagram "   " with layer "Business" in "diagram create blank/Motivation_Package"
    Then the create fails with "diagram name is required"

  Scenario: A duplicate diagram name in one package is refused
    Given the working model:
      | source | TestProject.xml        |
      | output | diagram_create_dup.xml |
      | root   | diagram create dup     |
    When I create a diagram "Dup" with layer "Business" in "diagram create dup/Motivation_Package"
    And I create a diagram "Dup" with layer "Business" in "diagram create dup/Motivation_Package"
    Then the create fails with "already has a diagram named"
