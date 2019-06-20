*** Settings ***
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***
${address}        P1HhWxfQLMgb5TfE56GASURCuitX2XL397G
${error_code}     -32602
${error_code2}    -32000
${result_code}    \f[a-z0-9]*
${result}         amounts is empty

*** Test Cases ***
Scenario: 20Contract - GetBalance
    [Tags]    normal
    [Template]    normalGetBalance
    ${address}
