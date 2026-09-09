Feature: Query failures are reported, never swallowed
  As an LLM agent
  I want a failed query to return a descriptive error and no partial result
  So that I learn why my statement did not run

  Background:
    Given a connection to the Sparx EA file "TestProject.eapx"

  Scenario Outline: Statements that cannot run return an error containing a reason
    When I run the query "<sql>"
    Then the query fails with "<message>"
    And no result is returned

    Examples:
      | sql                               | message                                        |
      | select * from t_not_a_real_table  | Got no result                                  |
      | selcet Object_ID from t_object    | Got no result                                  |
      | select No_Such_Column from t_object | Got no result                                  |
      | delete from t_object              | only read-only SELECT statements are supported |
      | update t_object set Name = 'x'    | only read-only SELECT statements are supported |
      | drop table t_object               | only read-only SELECT statements are supported |
      |                                   | empty SQL statement                            |

  Scenario: The error message quotes the failing statement
    When I run the query "select * from t_not_a_real_table"
    Then the query fails with "Got no result"
    And the query fails with "(sql: select * from t_not_a_real_table)"

  Scenario: Querying after Close fails instead of using a freed handle
    When I close the connection
    And I run the query "select Object_ID from t_object"
    Then the query fails with "connector is closed"
