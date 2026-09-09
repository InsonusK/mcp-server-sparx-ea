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

  Scenario: Copying a package into its own descendant is refused
    Given a new model:
      | root   | package copy cycle     |
      | output | package_copy_cycle.xml |
    When I create a package "Outer" in "package copy cycle"
    And I create a package "Inner" in "package copy cycle/Outer"
    And I copy the package "package copy cycle/Outer" into "package copy cycle/Outer/Inner"
    Then the create fails with "its own descendant"
