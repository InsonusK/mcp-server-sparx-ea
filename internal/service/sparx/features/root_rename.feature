Feature: Rename the model's root package (RootNodeNameChange)
  As the MCP server / a test author
  I want to rename the root package, optionally with a fresh identity
  So that many working copies can be imported into one Sparx EA project side by
  side (fresh identity) or a single model edited in place (keep identity)

  Scenario: The root of the read-only model is the EA root package
    Given the model file "TestProject.xml"
    When I read the model tree
    Then the root packages are "Model"

  Scenario: Renaming the root with a fresh identity — name, root id and every descendant GUID change
    Given the working model:
      | source   | TestProject.xml            |
      | output   | root_rename_fresh.xml      |
      | root     | ROOT renamed fresh         |
      | identity | fresh                      |
    When I read the element "ROOT renamed fresh/Motivation_Package/Goal1"
    Then the read succeeds
    And the element field "type" is "ArchiMate.Goal"
    And after reload the root package is "ROOT renamed fresh"
    And after reload the element "ROOT renamed fresh/Motivation_Package/Goal1" is:
      | field | value          |
      | type  | ArchiMate.Goal |
    And after reload the element "ROOT renamed fresh/Motivation_Package/Goal1" guid is not "{F3C2209E-B146-4328-863E-622BABD2B43E}"

  Scenario: Renaming the root while keeping identity — only the root changes, descendant GUIDs are untouched
    Given the working model:
      | source   | TestProject.xml       |
      | output   | root_rename_keep.xml  |
      | root     | ROOT renamed keep ids |
      | identity | keep                  |
    When I read the element "ROOT renamed keep ids/Motivation_Package/Goal1"
    Then the read succeeds
    And after reload the root package is "ROOT renamed keep ids"
    And after reload the element "ROOT renamed keep ids/Motivation_Package/Goal1" is:
      | field | value                                 |
      | guid  | {F3C2209E-B146-4328-863E-622BABD2B43E} |
      | type  | ArchiMate.Goal                        |
