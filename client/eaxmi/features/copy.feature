Feature: Copying a package subtree
  As sparx.AssembleReport and sparx.CopyPackage
  I want a deep copy with a fresh identity, that leaves references pointing out
  of the subtree alone
  So the same content can appear under another parent without sharing ids

  Scenario: A copy is deep and identity-disjoint; the source is untouched
    Given a new model with root "M"
    When I add a package "Source" under "M"
    And I add an element "G" of type "uml:Class" stereotype "ArchiMate_Goal" under "M/Source"
    And I add a package "Target" under "M"
    And I copy the package "M/Source" into "M/Target"
    And I write and reopen as "copy_deep.xml"
    Then resolving the element "M/Source/G" finds "G"
    And resolving the element "M/Target/Source/G" finds "G"
    And the reopened file has no dangling id references

  Scenario: A copied package keeps a connector whose far end is outside it
    Given a new model with root "M"
    When I add a package "Home" under "M"
    And I add an element "Anchor" of type "uml:Class" stereotype "ArchiMate_Goal" under "M/Home"
    And I add a package "Away" under "M"
    And I add an element "Pointer" of type "uml:Class" stereotype "ArchiMate_Requirement" under "M/Away"
    And I add a connector from "M/Away/Pointer" to "M/Home/Anchor" ea-type "Dependency" repr "dependency" stereotype "ArchiMate_Realization"
    And I copy the package "M/Away" into "M/Home"
    And I write and reopen as "copy_xref.xml"
    Then resolving the element "M/Home/Away/Pointer" finds "Pointer"
    And the connector from "M/Home/Away/Pointer" to "M/Home/Anchor" is:
      | field  | value      |
      | eaType | Dependency |
    And the reopened file has no dangling id references

  Scenario: Copying a whole package from another model brings its subtree
    Given a new model with root "Report"
    And a second model file "TestProject.xml"
    When I copy the package "Model/Motivation_Package" from the second model into "Report"
    And I write and reopen as "copy_crossdoc.xml"
    Then there are 15 elements
    And there are 7 connectors
    And resolving the element "Report/Motivation_Package/Goal1" finds "Goal1"
    And the reopened file has no dangling id references

  Scenario: Copying a package that does not exist is an error
    Given a new model with root "M"
    When I add a package "Real" under "M"
    Then copying the package "M/Ghost" into "M/Real" fails with "no package"
