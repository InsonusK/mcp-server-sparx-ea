Feature: Running SELECT queries against a Sparx EA file
  As an LLM agent
  I want Connector.Query to return exactly the rows and columns the SQL asks for
  So that I can read Sparx EA model data reliably

  Background:
    Given a connection to the Sparx EA file "TestProject.eapx"

  Scenario: Reading one element by id, including its memo (Note) column
    When I run the query "select Object_ID, Name, Object_Type, Stereotype, Author, Note from t_object where Object_ID = 7"
    Then the result is exactly:
      | Object_ID | Name  | Object_Type | Stereotype     | Author   | Note             |
      | 7         | Goal1 | Class       | ArchiMate_Goal | InsonusK | Goal description |

  Scenario: Reading the two packages
    When I run the query "select Package_ID, Name, Parent_ID from t_package"
    Then the result is exactly:
      | Package_ID | Name               | Parent_ID |
      | 1          | Model              | 0         |
      | 2          | Motivation_Package | 1         |

  Scenario: Reading relationships that end at the Goal element
    When I run the query "select Connector_ID, Connector_Type, Start_Object_ID, End_Object_ID from t_connector where End_Object_ID = 7"
    Then the result is exactly:
      | Connector_ID | Connector_Type | Start_Object_ID | End_Object_ID |
      | 6            | ControlFlow    | 4               | 7             |
      | 7            | Dependency     | 8               | 7             |
      | 8            | Dependency     | 10              | 7             |
      | 9            | Dependency     | 9               | 7             |
      | 10           | Dependency     | 11              | 7             |

  Scenario: Selecting all elements returns every row
    When I run the query "select Object_ID, Name from t_object"
    Then the result has 16 rows
    And the result contains the rows:
      | Object_ID | Name               |
      | 1         | Motivation_Package |
      | 2         | Stakeholder1       |
      | 7         | Goal1              |
      | 14        | ValueStream1       |
      | 16        | Location1          |

  Scenario: A filter that matches nothing yields an empty result, not an error
    When I run the query "select Object_ID, Name from t_object where Object_ID = 999999"
    Then the query succeeds
    And the result is empty
    And the result columns are "Object_ID, Name"

  Scenario Outline: The leading keyword is recognised regardless of letter case
    When I run the query "<sql>"
    Then the result is exactly:
      | Object_ID | Name  |
      | 7         | Goal1 |

    Examples:
      | sql                                                     |
      | select Object_ID, Name from t_object where Object_ID = 7 |
      | SELECT Object_ID, Name FROM t_object WHERE Object_ID = 7 |
      | SeLeCt Object_ID, Name from t_object where Object_ID = 7 |

  Scenario: Leading whitespace before the statement is tolerated
    When I run the query "    select Object_ID, Name from t_object where Object_ID = 7"
    Then the result is exactly:
      | Object_ID | Name  |
      | 7         | Goal1 |

  Scenario: The empty project has no elements but still has the root package
    Given a connection to the Sparx EA file "EmptyProject.eapx"
    When I run the query "select Object_ID, Name from t_object"
    Then the result is empty
    When I run the query "select Package_ID, Name, Parent_ID from t_package"
    Then the result is exactly:
      | Package_ID | Name  | Parent_ID |
      | 1          | Model | 0         |
