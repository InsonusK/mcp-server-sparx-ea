Feature: Concurrent access to one connection is serialized safely
  Architectural concern: the mdbtools engine is single-threaded. The Connector
  must guard it so that concurrent callers get correct results and no data race
  (the suite runs under -race).

  Scenario: Many goroutines share one connection
    Given a connection to the Sparx EA file "example/TestProject.eapx"
    When I run the query "select count(*) from t_object" from 40 goroutines concurrently
    Then all 40 queries succeed
    And every concurrent query returned the value "16"
