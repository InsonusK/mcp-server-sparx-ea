Feature: Queries run in-process
  Architectural concern (base-task DoD): the connector binds libmdbtools directly
  through cgo, so answering a query must not spawn a child process or write any
  temporary file.

  Scenario: Answering queries writes no temporary files
    Given an empty directory registered as the process temp dir
    And a connection to the Sparx EA file "example/TestProject.eapx"
    When I run the query "select Object_ID, Name from t_object" 25 times
    Then the temp dir is still empty
