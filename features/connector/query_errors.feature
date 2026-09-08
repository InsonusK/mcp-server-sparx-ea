Feature: Query errors are reported, not swallowed
  As an MCP server component
  I want a failed query to return a descriptive error
  So that an agent learns why its statement did not run

  Background:
    Given a connection to the Sparx EA file "example/TestProject.eapx"

  Scenario Outline: Rejected or failing statements
    When I run the query "<sql>"
    Then the query fails with an error containing "<message>"

    Examples:
      | sql                                    | message                                          |
      | select * from t_not_a_real_table       | Got no result                                    |
      | selcet Object_ID from t_object         | Got no result                                    |
      | select Bogus_Column from t_object      | Got no result                                    |
      | delete from t_object                    | only read-only SELECT statements are supported   |
      | update t_object set Name = 'x'          | only read-only SELECT statements are supported   |
      | drop table t_object                     | only read-only SELECT statements are supported   |
      |                                        | empty SQL statement                              |

  Scenario: Querying a closed connection fails
    When I close the connection
    And I run the query "select Object_ID from t_object"
    Then the query fails with an error containing "connector is closed"
