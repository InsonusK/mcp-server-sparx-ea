Feature: Rename the model's root package (RootNodeNameChange)
  As the MCP server / a test author
  I want the root package renamed, optionally with a fresh identity
  So that many working copies can be imported into one Sparx EA project side by
  side (fresh identity) or a single model edited in place (keep identity)

  The harness always renames the working copy's root to "NN <scenario name>"
  (unique per output file); the "identity" row of the working model chooses
  whether every descendant GUID is regenerated too.

  Scenario: The root of the read-only model is the EA root package
    Given the model file "TestProject.xml"
    When I read the model tree
    Then the root packages are "Model"

  Scenario: Renaming the root with a fresh identity regenerates every descendant GUID
    Given the working model:
      | source   | TestProject.xml       |
      | output   | root_rename_fresh.xml |
      | identity | fresh                 |
    When I read the element "Motivation_Package/Goal1"
    Then the read succeeds
    And the element field "type" is "ArchiMate.Goal"
    And after reload the root package is this scenario's package
    And after reload the element "Motivation_Package/Goal1" is:
      | field | value          |
      | type  | ArchiMate.Goal |
    And after reload the element "Motivation_Package/Goal1" guid is not "{F3C2209E-B146-4328-863E-622BABD2B43E}"

  Scenario: Renaming the root while keeping identity leaves descendant GUIDs untouched
    Given the working model:
      | source   | TestProject.xml      |
      | output   | root_rename_keep.xml |
      | identity | keep                 |
    When I read the element "Motivation_Package/Goal1"
    Then the read succeeds
    And after reload the root package is this scenario's package
    And after reload the element "Motivation_Package/Goal1" is:
      | field | value                                 |
      | guid  | {F3C2209E-B146-4328-863E-622BABD2B43E} |
      | type  | ArchiMate.Goal                        |
