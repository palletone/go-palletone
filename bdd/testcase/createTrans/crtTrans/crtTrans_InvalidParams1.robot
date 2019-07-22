*** Settings ***
Force Tags        invalidAdd
Default Tags      invalidAdd
Library           RequestsLibrary
Library           Collections
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         wallet_CreateRawTransaction
${error_code}     -32602
${error_message}    missing value for required argument 3
${error_message2}    too many arguments, want at most 4

*** Test Cases ***
invalidParams1
    [Tags]    invalidParams
    ${crtList1}    Given I set Params which is less than required
    ${resp1}    When I post a crtTrans request    ${crtList1}
    Then I get a code and a message    ${resp1}    ${error_code}    ${error_message}

*** Keywords ***
