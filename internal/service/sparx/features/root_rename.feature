Feature: Rename the model's root package (RootNodeNameChange)
  As the MCP server / a test author
  I want each working copy to have a uniquely named root package with a fresh id
  So that many working copies can be imported into one Sparx EA project side by side
  and compared against the original

  Scenario: The root of the read-only model is the EA root package
    Given the model file "TestProject.xml"
    When I read the model tree
    Then the root packages are "Model"

  Scenario: Renaming the root changes the name and the identity, and paths follow
    Given the working model:
      | source | TestProject.xml            |
      | output | root_rename.xml            |
      | root   | ROOT was renamed by a test |
    When I read the element "ROOT was renamed by a test/Motivation_Package/Goal1"
    Then the read succeeds
    And the element field "type" is "ArchiMate.Goal"
    And after reload the root package is "ROOT was renamed by a test"
    And after reload the element "ROOT was renamed by a test/Motivation_Package/Goal1" is:
      | field | value          |
      | type  | ArchiMate.Goal |
