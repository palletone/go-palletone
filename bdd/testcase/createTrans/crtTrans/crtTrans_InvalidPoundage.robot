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
poundageInvalid1
    [Tags]    invalidAmount
    [Template]    setInvalidPoundage
    r443    -32602    invalid argument 3: Error decoding string
    4fgf    -32602    invalid argument 3: Error decoding string
    ${SPACE}    -32602    invalid argument 3: Error decoding string
    $    -32602    invalid argument 3: Error decoding string
    "0"    -32602    invalid argument 3: Error decoding string

poundageInvalid2
    [Tags]    invalidAmount
    [Template]    setInvalidPoundage
    0    -32000    fee is invalid
    -4.5    -32000    fee is invalid

poundageInvalid3
    [Tags]    invalidAmount
    [Template]    setInvalidPoundage
    10000000000001    -32000    Select utxo err

poundageEmpty
    [Tags]    invalidAmount
    [Template]    setInvalidPoundage
    ${Empty}    -32602    invalid argument 3: Error decoding string

*** Keywords ***
