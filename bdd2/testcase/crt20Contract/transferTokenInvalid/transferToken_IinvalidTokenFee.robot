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
Scenario: InvalidTokenFee
    [Template]    InvalidTransferToken
    ${tokenId}    2    100000000000    description    1    ${6000000}    -32000
    ...    Select PTN utxo err    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    2    -1.00000001    description    1    ${6000000}    -32000
    ...    fee is ZERO    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    2    -0.00000001    description    1    ${6000000}    -32000
    ...    fee is ZERO    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    2    %    description    1    ${6000000}    -32602
    ...    invalid argument 4: Error decoding string    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    2    j    description    1    ${6000000}    -32602
    ...    invalid argument 4: Error decoding string    ${listAccounts[0]}    ${listAccounts[1]}
    ${tokenId}    2    ${EMPTY}    description    1    ${6000000}    -32602
    ...    invalid argument 4: Error decoding string    ${listAccounts[0]}    ${listAccounts[1]}

*** Keywords ***
