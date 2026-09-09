Feature: Read diagram contents and layout (method 3)

  Background:
    Given the model file "TestProject.xml"

  Scenario: A diagram reports its type and its placed elements with coordinates
    When I read the diagram "Model/Motivation_Package/Motivation_Diagram"
    Then the diagram field "diagramType" is "Logical"
    And the diagram has 15 placed elements
    And the diagram has 7 links
    And the placed elements include:
      | name         | type                  | left | top | right | bottom |
      | Location1    | ArchiMate.Location    | 520  | 680 | 610   | 750    |
      | Goal1        | ArchiMate.Goal        | 460  | 250 | 560   | 320    |
      | Stakeholder1 | ArchiMate.Stakeholder | 70   | 60  | 170   | 130    |

  Scenario: The same diagram is reachable by GUID
    When I read the diagram "{EA3BF442-E212-4e5e-9812-23F14C78B4BE}"
    Then the diagram field "name" is "Motivation_Diagram"

  Scenario: Read a Cyrillic diagram
    Given the model file "CyrillicProject.xml"
    When I read the diagram "Модель/Пакет/Диаграмма"
    Then the diagram has 1 placed elements
    And the placed elements include:
      | name |
      | Цель |
