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
Scenario: InvalidRecieverAdd
    [Template]    InvalidTransferToken
    ${tokenId}    2    1    description    1    ${6000000}    -32000
    ...    Address cannot be null    ${listAccounts[0]}    ${EMPTY}
    ${tokenId}    2    1    description    1    ${6000000}    -32000
    ...    PalletOne address must start with 'P'    ${listAccounts[0]}    p
    ${tokenId}    2    1    description    1    ${6000000}    -32000
    ...    invalid format: version and/or checksum bytes missing    ${listAccounts[0]}    P1
    ${tokenId}    2    1    description    1    ${6000000}    -32000
    ...    checksum error    ${listAccounts[0]}    P1HhWxfQLMgb5TfE56GASURCuitX2XL397
    ${tokenId}    2    1    description    1    ${6000000}    -32000
    ...    checksum error    ${listAccounts[0]}    P1HhWxfQLMgb5TfE56GASURCuitX2XL397G1

*** Keywords ***
