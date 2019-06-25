*** Settings ***
Library           Collections
Library           BuiltIn

*** Test Cases ***
test
    ${dict}=    Create Dictionary    a    1    b    2
    ${status}    ${value}=    Run Keyword And Ignore Error    Get From Dictionary    ${dict}    result
    Run Keyword If    '${status}' == 'PASS'    Log    "111"
