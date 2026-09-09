Feature: Copy a package (deep, with a fresh identity)
  As the MCP server / the report assembler
  I want to duplicate a package subtree elsewhere in the model
  So that the same content can appear under another parent without sharing ids

  Scenario: Copy a package within the same root node
    Given a new model:
      | root   | package copy same    |
      | output | package_copy_same.xml |
    When I create a package "Source" in "package copy same"
    And I create a "ArchiMate.Goal" named "G1" in "package copy same/Source" with note "keep me"
    And I create a package "Target" in "package copy same"
    And I copy the package "package copy same/Source" into "package copy same/Target"
    Then the create succeeds
    And after reload the package "package copy same/Source" elements are "G1"
    And after reload the package "package copy same/Target/Source" elements are "G1"
    And after reload the element "package copy same/Target/Source/G1" is:
      | field         | value          |
      | type          | ArchiMate.Goal |
      | documentation | keep me        |

  Scenario: Copy a package between two root nodes
    Given a new model:
      | root   | package copy origin     |
      | output | package_copy_between.xml |
    When I create a root package "package copy destination"
    And I create a package "Shared" in "package copy origin"
    And I create a "ArchiMate.Requirement" named "R1" in "package copy origin/Shared" with note ""
    And I copy the package "package copy origin/Shared" into "package copy destination"
    Then the create succeeds
    And after reload the package "package copy origin/Shared" elements are "R1"
    And after reload the package "package copy destination/Shared" elements are "R1"
    And after reload the package "package copy destination/Shared" is:
      | field  | value                    |
      | parent | package copy destination |

  Scenario: A copied package keeps a relationship that points outside it
    Given a new model:
      | root   | package copy xref     |
      | output | package_copy_xref.xml |
    When I create a package "Anchor" in "package copy xref"
    And I create a "ArchiMate.Goal" named "AnchorGoal" in "package copy xref/Anchor" with note ""
    And I create a package "Mobile" in "package copy xref"
    And I create a "ArchiMate.Requirement" named "MobileReq" in "package copy xref/Mobile" with note ""
    And I relate "package copy xref/Mobile/MobileReq" to "package copy xref/Anchor/AnchorGoal" as "ArchiMate.Realization"
    And I copy the package "package copy xref/Mobile" into "package copy xref/Anchor"
    Then the create succeeds
    # the copy's connector still resolves to the original AnchorGoal (a foreign ref,
    # left alone like EA does in example/side1.xml)
    And after reload the element "package copy xref/Anchor/Mobile/MobileReq" relations include:
      | type                  | direction | otherName  |
      | ArchiMate.Realization | outgoing  | AnchorGoal |

  Scenario: A copied package keeps a diagram object that shows an element outside it
    Given the working model:
      | source | TestProject.xml            |
      | output | package_copy_xref_diag.xml |
      | root   | package copy xref diagram  |
    When I create a package "Outside" in "package copy xref diagram"
    And I create a "ArchiMate.Goal" named "Visitor" in "package copy xref diagram/Outside" with note ""
    And I add "package copy xref diagram/Outside/Visitor" to the diagram "package copy xref diagram/Motivation_Package/Motivation_Diagram" at 10,10,110,80
    And I copy the package "package copy xref diagram/Motivation_Package" into "package copy xref diagram/Outside"
    Then the create succeeds
    And after reload the diagram "package copy xref diagram/Outside/Motivation_Package/Motivation_Diagram" has 16 placed elements
    And after reload the diagram "package copy xref diagram/Outside/Motivation_Package/Motivation_Diagram" placed elements include:
      | name    |
      | Visitor |
      | Goal1   |

  Scenario: Copying a package into its own descendant is refused
    Given a new model:
      | root   | package copy cycle     |
      | output | package_copy_cycle.xml |
    When I create a package "Outer" in "package copy cycle"
    And I create a package "Inner" in "package copy cycle/Outer"
    And I copy the package "package copy cycle/Outer" into "package copy cycle/Outer/Inner"
    Then the create fails with "its own descendant"
