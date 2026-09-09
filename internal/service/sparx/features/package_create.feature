Feature: Create a package

  Each scenario declares its fixture and its own tmp file.

  Scenario: Create a sub-package
    Given the working model:
      | source | TestProject.xml    |
      | output | package_create.xml |
      | root   | package create     |
    When I create a package "Realisation" in "package create/Motivation_Package"
    Then the create succeeds
    And after reload the package "package create/Motivation_Package" packages are "Realisation"
    And after reload the package "package create/Motivation_Package/Realisation" is:
      | field  | value                                        |
      | path   | package create/Motivation_Package/Realisation |
      | parent | package create/Motivation_Package            |
    And after reload the package "package create/Motivation_Package/Realisation" elements are ""

  Scenario: Create a nested package then an element inside it
    Given the working model:
      | source | TestProject.xml           |
      | output | package_create_nested.xml |
      | root   | package create nested     |
    When I create a package "Sub" in "package create nested/Motivation_Package"
    And I create a "ArchiMate.Goal" named "NestedGoal" in "package create nested/Motivation_Package/Sub" with note ""
    Then the create succeeds
    And after reload the package "package create nested/Motivation_Package/Sub" elements are "NestedGoal"

  Scenario: A duplicate sub-package name is refused
    Given the working model:
      | source | TestProject.xml        |
      | output | package_create_dup.xml |
      | root   | package create dup     |
    When I create a package "Motivation_Package" in "package create dup"
    Then the create fails with "already has a sub-package named"

  Scenario: A blank package name is refused
    Given the working model:
      | source | TestProject.xml          |
      | output | package_create_blank.xml |
      | root   | package create blank     |
    When I create a package "   " in "package create blank/Motivation_Package"
    Then the create fails with "package name is required"
