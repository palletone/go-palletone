*** Settings ***
Library           RequestsLibrary
Library           Collections
Library           demjson
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/utilDefined.txt

*** Variables ***
${host}           http://localhost:8545/
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
