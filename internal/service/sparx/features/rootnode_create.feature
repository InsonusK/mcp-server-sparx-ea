Feature: Create a model from scratch (a new EA root node)
  As the MCP server
  I want to build a model that is not a copy of an existing export
  So that a report / a new project can start empty

  Scenario: A new empty model has just its root package
    Given a new model:
      | root   | rootnode create empty |
      | output | rootnode_create.xml   |
    Then the root packages are "rootnode create empty"
    And after reload the root package is "rootnode create empty"

  Scenario: Create a package directly in the root node
    Given a new model:
      | root   | rootnode create with package |
      | output | rootnode_create_package.xml  |
    When I create a package "Sub" in "rootnode create with package"
    Then the create succeeds
    And after reload the package "rootnode create with package/Sub" is:
      | field  | value                        |
      | name   | Sub                          |
      | path   | rootnode create with package/Sub |
      | parent | rootnode create with package |

  Scenario: Build a nested package tree from scratch
    Given a new model:
      | root   | rootnode create tree      |
      | output | rootnode_create_tree.xml  |
    When I create a package "A" in "rootnode create tree"
    And I create a package "B" in "rootnode create tree/A"
    Then the create succeeds
    And after reload the package "rootnode create tree" packages are "A"
    And after reload the package "rootnode create tree/A" packages are "B"
