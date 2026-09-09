Feature: Queries run in-process with no side effects on disk
  Architectural concern (base-task DoD): the connector binds libmdbtools through
  cgo. Reading must not modify the .eapx file and must not write temp files.

  Scenario: Repeated reads leave the file and the temp dir untouched
    Given the on-disk state of "TestProject.eapx" is recorded
    And an empty directory is set as TMPDIR
    And a connection to the Sparx EA file "TestProject.eapx"
    When I run the query "select Object_ID, Name, Note from t_object" 20 times
    Then the "TestProject.eapx" file on disk is unchanged
    And TMPDIR is still empty
