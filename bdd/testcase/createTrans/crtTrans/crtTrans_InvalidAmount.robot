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
${method}         ptn_cmdCreateTransaction

*** Test Cases ***
amountInvalid1
    [Tags]    invalidAmount
    [Template]    setInvalidAmount
    ferew    -32602    invalid argument 2: Error decoding string

amountInvalid2
    [Tags]    invalidAmount
    [Template]    setInvalidAmount
    3d    -32602    invalid argument 2: Error decoding string

amountInvalid3
    [Tags]    invalidAmount
    [Template]    setInvalidAmount
    $    -32602    invalid argument 2: Error decoding string

amountInvalid4
    [Tags]    invalidAmount
    [Template]    setInvalidAmount
    0    -32000    amounts is invalid

amountInvalid5
    [Tags]    invalidAmount
    [Template]    setInvalidAmount
    0    -32000    amounts is invalid
    -6    -32000    amounts is invalid

amountEmpty
    [Tags]    invalidAmount
    [Template]    setInvalidAmount
    ${Empty}    -32602    invalid argument 2: Error decoding string

*** Keywords ***
