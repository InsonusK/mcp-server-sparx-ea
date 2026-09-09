Feature: Parsing an EA "Export Package to XMI" file
  As internal/service/sparx
  I want the raw XMI turned into plain structs — packages, elements, connectors,
  diagrams — with the EA-specific <xmi:Extension> data folded in
  So the service can work with the model without touching XML

  Scenario: A populated model reports its shape
    Given the model file "TestProject.xml"
    Then the exporter version is "6.5"
    And the root package names are "Model"
    And there are 15 elements
    And there are 7 connectors
    And there are 2 packages

  Scenario: An element carries its ArchiMate stereotype, uml type and note
    Given the model file "TestProject.xml"
    Then the element "Model/Motivation_Package/Goal1" is:
      | field         | value                                    |
      | id            | EAID_F3C2209E_B146_4328_863E_622BABD2B43E |
      | guid          | {F3C2209E-B146-4328-863E-622BABD2B43E}    |
      | name          | Goal1                                    |
      | umlType       | uml:Class                               |
      | stereotype    | ArchiMate_Goal                          |
      | documentation | Goal description                        |
      | package       | Model/Motivation_Package                |

  Scenario: A connector keeps its EA type, stereotype and endpoints
    Given the model file "TestProject.xml"
    Then the connector from "Model/Motivation_Package/Driver1" to "Model/Motivation_Package/Assessment1" is:
      | field      | value                 |
      | name       | Driver to Assessment  |
      | eaType     | ControlFlow           |
      | stereotype | ArchiMate_Influence   |

  Scenario: An element with no stereotype keeps an empty stereotype, not a made-up one
    Given the model file "PlainUML.xml"
    Then the element "Plain/PlainClass" is:
      | field      | value     |
      | umlType    | uml:Class |
      | stereotype |           |

  Scenario: A diagram lists its placed elements with coordinates
    Given the model file "DeepModel.xml"
    Then the diagram "DeepModel/Package1/Package2/Package3/Package3Diagram" has 2 objects and 1 links
    And the diagram "DeepModel/Package1/Package2/Package3/Package3Diagram" places "DeepModel/Package1/Package2/Package3/Goal1" at 330,80,430,150

  Scenario: Cyrillic names and notes round-trip without mojibake
    Given the model file "CyrillicProject.xml"
    Then the root package names are "Модель"
    And the element "Модель/Пакет/Цель" is:
      | field         | value          |
      | name          | Цель           |
      | documentation | Описание Цели   |
    And every element name and documentation is valid UTF-8

  Scenario: Deep package nesting is parsed to any depth
    Given the model file "DeepModel.xml"
    Then there are 4 packages
    And the element "DeepModel/Package1/Package2/Package3/Goal1" is:
      | field   | value                                    |
      | package | DeepModel/Package1/Package2/Package3      |
