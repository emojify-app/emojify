Feature: test emojify get
  In order to test the emojify service
  As a developer
  I need to test the GET interface

  Scenario: get url
    Given the server is running
    And i post an image url to the endpoint
    When i call get with the image url
    Then i expect a valid image to be returned
