Feature: Finding and repairing EA-representation mismatches (method 4)
  Architectural concern: a model exported before eaForElement (archimate.go)
  learned the right EA base type for every ArchiMate type still has elements
  stored with the wrong one — Sparx EA renders these as anonymous classes
  instead of their real ArchiMate shape. ValidateModel finds them,
  FixElements repairs the ones it's given, ValidateAndFixModel does both in
  one step. testdata/LegacyEAIssues.xml is a small hand-built fixture shaped
  like a pre-fix export: three broken elements and two already-correct ones.

  Scenario: Validate reports exactly the elements with the wrong EA base type
    Given the model file "LegacyEAIssues.xml"
    When I validate the model
    Then the validation reports exactly these issues:
      | name           | type                           | gotType   | wantType      |
      | OldInterface   | ArchiMate.TechnologyInterface  | uml:Class | uml:Interface |
      | OldComponent   | ArchiMate.ApplicationComponent | uml:Class | uml:Component |
      | OldValueStream | ArchiMate.ValueStream          | uml:Class | uml:Activity  |

  Scenario: Fix repairs the elements it's given and leaves the rest untouched
    Given the working model:
      | source | LegacyEAIssues.xml       |
      | output | element_ea_repair_fix.xml |
      | root   | ea repair fix            |
    When I fix the elements:
      | ea repair fix/Motivation_Package/OldInterface |
      | ea repair fix/Motivation_Package/OldComponent |
    Then the fix succeeds
    And after reload the element "ea repair fix/Motivation_Package/OldInterface" has EA type "uml:Interface"
    And after reload the element "ea repair fix/Motivation_Package/OldComponent" has EA type "uml:Component"

  Scenario: Fix is a no-op for an already-correct element or an unknown ref
    Given the working model:
      | source | LegacyEAIssues.xml         |
      | output | element_ea_repair_noop.xml |
      | root   | ea repair noop             |
    When I fix the elements:
      | ea repair noop/Motivation_Package/GoodGoal |
      | ea repair noop/Motivation_Package/Ghost    |
    Then the fix succeeds
    And the fix reports 0 issues
    And after reload the element "ea repair noop/Motivation_Package/GoodGoal" has EA type "uml:Class"

  Scenario: Validate-and-fix repairs every mismatch in one step
    Given the working model:
      | source | LegacyEAIssues.xml        |
      | output | element_ea_repair_all.xml |
      | root   | ea repair all             |
    When I validate and fix the model
    Then the fix succeeds
    And the fix reports 3 issues
    And after reload the element "ea repair all/Motivation_Package/OldInterface" has EA type "uml:Interface"
    And after reload the element "ea repair all/Motivation_Package/OldComponent" has EA type "uml:Component"
    And after reload the element "ea repair all/Motivation_Package/OldValueStream" has EA type "uml:Activity"
    And after reload the element "ea repair all/Motivation_Package/GoodGoal" has EA type "uml:Class"
    And after reload the element "ea repair all/Motivation_Package/GoodProcess" has EA type "uml:Activity"
