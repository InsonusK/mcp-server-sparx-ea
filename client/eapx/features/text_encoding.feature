Feature: Text columns are decoded to valid UTF-8
  Architectural concern (base-task DoD): text and memo columns stored in the
  JET database (UCS-2LE for JET4, CP1251 for legacy JET3) must reach the caller
  as valid UTF-8 without mojibake.

  Scenario: Every text value returned by a query is valid UTF-8
    Given a connection to the Sparx EA file "example/TestProject.eapx"
    When I run the query "select Name, Note, Object_Type, Stereotype from t_object"
    Then the query succeeds
    And every result value is valid UTF-8

  # Excluded from every run via the ~@todo tag filter (see the godog runners).
  # Blocked: authoring a Cyrillic .eapx needs Sparx EA itself — mdbtools is
  # read-only and cannot INSERT rows, and neither sample project contains
  # non-ASCII text. Add example/CyrillicProject.eapx and drop the @todo tag to
  # enable it. Tracked in docs/test-trace-matrix.md.
  @todo
  Scenario: Cyrillic text round-trips without mojibake
    Given a connection to the Sparx EA file "example/CyrillicProject.eapx"
    When I run the query "select Name from t_object where Object_ID = 1"
    Then the single result value is "Кириллица объекта"
