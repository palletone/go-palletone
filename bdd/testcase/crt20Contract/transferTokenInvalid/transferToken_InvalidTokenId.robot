*** Settings ***
Suite Setup       preTransToken
Default Tags      invalid
Library           RequestsLibrary
Library           Collections
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         wallet_transferToken

*** Test Cases ***
Scenario: InvalidTokenId
    [Template]    InvalidTransferToken
    ${EMPTY}    2    1    description    1    ${6000000}    -32000
    ...    Asset string invalid    ${listAccounts[0]}    ${listAccounts[1]}
    &    2    1    description    1    ${6000000}    -32000
    ...    Asset string invalid    ${listAccounts[0]}    ${listAccounts[1]}
    fr    2    1    description    1    ${6000000}    -32000
    ...    Asset string invalid    ${listAccounts[0]}    ${listAccounts[1]}
    63    2    1    description    1    ${6000000}    -32000
    ...    Asset string invalid    ${listAccounts[0]}    ${listAccounts[1]}
    'QA001+104P0SBEQ47Z8L7ELOV '    2    1    description    1    ${6000000}    -32000
    ...    Symbol must less than 5 characters    ${listAccounts[0]}    ${listAccounts[1]}
    QA0012+104P0SBEQ47Z8L7ELOV    2    1    description    1    ${6000000}    -32000
    ...    Symbol must less than 5 characters    ${listAccounts[0]}    ${listAccounts[1]}
    QA001+104P0SBEQ47Z8L7ELO V    2    1    description    1    ${6000000}    -32000
    ...    requestId must more than 10    ${listAccounts[0]}    ${listAccounts[1]}

*** Keywords ***
