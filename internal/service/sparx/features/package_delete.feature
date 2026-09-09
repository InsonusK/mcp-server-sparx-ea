Feature: Delete a package

  DeletePackage has a cascadeDelete flag: off (default) it refuses a package
  that still has contents; on, it removes the whole subtree.

  Scenario: Delete an empty package
    Given the working model:
      | source | TestProject.xml    |
      | output | package_delete.xml |
      | root   | package delete     |
    When I create a package "Spare" in "package delete/Motivation_Package"
    And I delete the package "package delete/Motivation_Package/Spare"
    Then the delete succeeds
    And after reload the package "package delete/Motivation_Package/Spare" cannot be found
    And after reload the package "package delete/Motivation_Package" packages are ""

  Scenario: Deleting a non-empty package without cascade is refused
    Given the working model:
      | source | TestProject.xml            |
      | output | package_delete_nonempty.xml |
      | root   | package delete nonempty     |
    When I create a package "Doomed" in "package delete nonempty/Motivation_Package"
    And I create a "ArchiMate.Goal" named "Inside" in "package delete nonempty/Motivation_Package/Doomed" with note ""
    And I delete the package "package delete nonempty/Motivation_Package/Doomed"
    Then the delete fails with "not empty"

  Scenario: Deleting a non-empty package with its contents removes the subtree
    Given the working model:
      | source | TestProject.xml             |
      | output | package_delete_cascade.xml  |
      | root   | package delete cascade      |
    When I create a package "Doomed" in "package delete cascade/Motivation_Package"
    And I create a "ArchiMate.Goal" named "Inside" in "package delete cascade/Motivation_Package/Doomed" with note ""
    And I delete the package "package delete cascade/Motivation_Package/Doomed" with its contents
    Then the delete succeeds
    And after reload the package "package delete cascade/Motivation_Package/Doomed" cannot be found
    And after reload the element "package delete cascade/Motivation_Package/Doomed/Inside" cannot be found

  Scenario: Deleting an EA root package is refused
    Given the working model:
      | source | TestProject.xml         |
      | output | package_delete_root.xml |
      | root   | package delete root     |
    When I delete the package "package delete root"
    Then the delete fails with "root package"
