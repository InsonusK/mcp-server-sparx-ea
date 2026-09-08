Feature: Run SQL queries against a Sparx EA file
  As an MCP server component
  I want to send SQL SELECT statements to a .eapx file
  So that an agent can read model data out of it

  Background:
    Given a connection to the Sparx EA file "example/TestProject.eapx"

  Scenario Outline: Selecting rows from the model
    When I run the query "<sql>"
    Then the query succeeds
    And the result has <rows> rows
    And the result columns are "<columns>"

    Examples:
      | sql                                                                       | rows | columns                        |
      | select Object_ID, Name, Object_Type from t_object                          | 16   | Object_ID,Name,Object_Type     |
      | select Object_ID from t_object where Object_ID = 999999                     | 0    | Object_ID                      |
      | select count(*) from t_object                                              | 1    | count                          |
      | select Name, Notes from t_package                                          | 2    | Name,Notes                     |

  Scenario: Reading a specific element by id
    When I run the query "select Object_ID, Name, Object_Type from t_object where Object_ID = 2"
    Then the query succeeds
    And the result has 1 rows
    And row 0 equals "2,Stakeholder1,Class"

  Scenario: Aggregate query returns the computed value
    When I run the query "select count(*) from t_object"
    Then the query succeeds
    And the single result value is "16"

  Scenario: An empty project yields no model elements
    Given a connection to the Sparx EA file "example/EmptyProject.eapx"
    When I run the query "select Object_ID, Name from t_object"
    Then the query succeeds
    And the result has 0 rows
