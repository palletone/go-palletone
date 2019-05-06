*** Settings ***
Suite Setup       preTransToken
Default Tags      invalidAdd
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
Scenario: InvalidTokenAmount
    [Template]    InvalidTransferToken
    ${tokenId}    -2.1    1    description    1    ${6000000}    -32000
    ...    Select utxo err    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    10000000000    1    description    1    ${6000000}    -32000
    ...    Select utxo err    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    %    1    description    1    ${6000000}    -32602
    ...    invalid argument 3: Error decoding string    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    j    1    description    1    ${6000000}    -32602
    ...    invalid argument 3: Error decoding string    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    ${EMPTY}    1    description    1    ${6000000}    -32602
    ...    invalid argument 3: Error decoding string    ${listAccounts[0]}    ${listAccounts[1]}

*** Keywords ***
