Feature: test emojify post
  In order to test the emojify service
  As a developer
  I need to test the POST interface

  Scenario: post url
    Given the server is running
    When i post an image url to the endpoint
    Then i expect a base64 url to be returned
