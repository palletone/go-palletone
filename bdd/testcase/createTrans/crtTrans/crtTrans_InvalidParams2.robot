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
${error_message}    too many arguments, want at most 4

*** Test Cases ***
invalidParams2
    [Tags]    invalidParams
    ${crtList2}    Given I set Params which is more than required
    ${resp2}    When I_post_a_crtTrans_request    ${crtList2}
    Then I get a code and a message    ${resp2}    ${error_code}    ${error_message}

*** Keywords ***
