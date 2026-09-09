Feature: Resolving refs and walking the model
  As internal/service/sparx
  I want to turn an id, a braced GUID or a slash path into a model object
  So the service can accept whichever the caller has

  Scenario Outline: A ref is recognised as an id or a path
    Then "<ref>" <verb> a ref id

    Examples:
      | ref                                      | verb   |
      | EAID_F3C2209E_B146_4328_863E_622BABD2B43E | is     |
      | EAPK_12EDEB75_D31A_4b24_9476_EA75A66E41E8 | is     |
      | {F3C2209E-B146-4328-863E-622BABD2B43E}    | is     |
      | Model/Motivation_Package/Goal1            | is not |
      | Goal1                                     | is not |

  Scenario: A path splits into its trimmed, non-empty segments
    Then splitting the path "  Model / Motivation_Package // Goal1 " gives "Model, Motivation_Package, Goal1"

  Scenario: An element resolves by path, by id and by GUID
    Given the model file "TestProject.xml"
    When resolving the element "Model/Motivation_Package/Goal1" finds "Goal1"
    Then the same element resolves by id and by GUID

  Scenario: A missing path resolves to nothing, not an error
    Given the model file "TestProject.xml"
    Then resolving the element "Model/Motivation_Package/DoesNotExist" finds nothing

  Scenario: Diagrams and packages resolve by path
    Given the model file "TestProject.xml"
    Then resolving the diagram "Model/Motivation_Package/Motivation_Diagram" finds "Motivation_Diagram"
    And resolving the package "Model/Motivation_Package" finds "Motivation_Package"

  Scenario: Paths reflect the real nesting
    Given the model file "DeepModel.xml"
    Then the path of the element "DeepModel/Package1/Package2/Package3/Goal1" is "DeepModel/Package1/Package2/Package3/Goal1"
    And the package paths are "DeepModel, DeepModel/Package1, DeepModel/Package1/Package2, DeepModel/Package1/Package2/Package3"

  Scenario: Every element in the model is enumerable
    Given the model file "DeepModel.xml"
    Then the element names are "Goal1, Resource1"
