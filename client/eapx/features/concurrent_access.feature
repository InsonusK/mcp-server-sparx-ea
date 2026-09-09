Feature: Concurrent callers of one connection are serialized
  Architectural concern: the mdbtools engine is single-threaded and keeps cursor
  state on the handle. The Connector mutex must make concurrent Query calls
  behave as if run one at a time. If the mutex is removed, goroutines querying
  different objects see each other's rows and this test fails. Runs under -race.

  Background:
    Given a connection to the Sparx EA file "TestProject.eapx"

  Scenario: 60 goroutines each read a different element and get exactly its row
    Given the elements to read concurrently:
      | Object_ID | Name         | Object_Type |
      | 2         | Stakeholder1 | Class       |
      | 7         | Goal1        | Class       |
      | 14        | ValueStream1 | Activity    |
    When 60 goroutines read those elements concurrently, round-robin
    Then no goroutine returned an error
    And every goroutine got exactly its assigned row
