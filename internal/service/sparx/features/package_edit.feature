Feature: Rename and move a package

  Each scenario declares its fixture and its own tmp file.

  Scenario: Rename a package
    Given the working model:
      | source | TestProject.xml  |
      | output | package_edit.xml |
      | root   | package edit     |
    When I rename the package "package edit/Motivation_Package" to "Motivation"
    Then the rename succeeds
    And after reload the package "package edit/Motivation" is:
      | field | value      |
      | name  | Motivation |
    And after reload the element "package edit/Motivation/Goal1" is:
      | field | value          |
      | type  | ArchiMate.Goal |

  Scenario: Renaming to a blank name is refused
    Given the working model:
      | source | TestProject.xml      |
      | output | package_edit_bad.xml |
      | root   | package edit bad     |
    When I rename the package "package edit bad/Motivation_Package" to "   "
    Then the rename fails with "new name is required"

  Scenario: Move a package under another package
    Given the working model:
      | source | TestProject.xml  |
      | output | package_move.xml |
      | root   | package move     |
    When I create a package "Outer" in "package move/Motivation_Package"
    And I create a package "Inner" in "package move/Motivation_Package"
    And I move the package "package move/Motivation_Package/Inner" into "package move/Motivation_Package/Outer"
    Then the move succeeds
    And after reload the package "package move/Motivation_Package/Outer" packages are "Inner"
    And after reload the package "package move/Motivation_Package" packages are "Outer"
    And after reload the package "package move/Motivation_Package/Outer/Inner" is:
      | field  | value                                    |
      | parent | package move/Motivation_Package/Outer     |

  Scenario: Moving a package into its own descendant is refused
    Given the working model:
      | source | TestProject.xml          |
      | output | package_move_cycle.xml   |
      | root   | package move cycle       |
    When I create a package "Parent" in "package move cycle/Motivation_Package"
    And I create a package "Child" in "package move cycle/Motivation_Package/Parent"
    And I move the package "package move cycle/Motivation_Package/Parent" into "package move cycle/Motivation_Package/Parent/Child"
    Then the move fails with "own descendant"
