Feature: Text is decoded from the JET database to UTF-8
  Architectural concern (base-task DoD): Name/Note text stored in the .eapx
  (UCS-2LE for JET4) must reach the caller as valid UTF-8 without mojibake,
  including Cyrillic.

  Scenario: Cyrillic Name and Note round-trip exactly
    Given a connection to the Sparx EA file "CyrillicProject.eapx"
    When I run the query "select Object_ID, Name, Object_Type, Note from t_object where Object_ID = 3"
    Then the result is exactly:
      | Object_ID | Name | Object_Type | Note          |
      | 3         | Цель | Class       | Описание Цели  |

  Scenario: Cyrillic packages round-trip exactly
    Given a connection to the Sparx EA file "CyrillicProject.eapx"
    When I run the query "select Package_ID, Name from t_package"
    Then the result is exactly:
      | Package_ID | Name   |
      | 2          | Модель |
      | 3          | Пакет  |

  Scenario: A Cyrillic string literal works in the WHERE clause
    Given a connection to the Sparx EA file "CyrillicProject.eapx"
    When I run the query "select Object_ID, Name from t_object where Name = 'Цель'"
    Then the result is exactly:
      | Object_ID | Name |
      | 3         | Цель |

  Scenario: The decoded Cyrillic value is the exact UTF-8 byte sequence
    Given a connection to the Sparx EA file "CyrillicProject.eapx"
    When I run the query "select Name from t_object where Object_ID = 3"
    Then the single result cell is the UTF-8 bytes "d0a6d0b5d0bbd18c"

  Scenario: Every text value from a query over the populated project is valid UTF-8
    Given a connection to the Sparx EA file "TestProject.eapx"
    When I run the query "select Name, Note, Object_Type, Stereotype, Author from t_object"
    Then the query succeeds
    And every result value is valid UTF-8
