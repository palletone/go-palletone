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
recieverInvalid1
    [Tags]    invalidAdd1
    [Template]    setInvalidReciever
    P1    -32000    receiver address is invalid

recieverInvalid2
    [Tags]    invalidAdd2
    [Template]    setInvalidReciever
    P    -32000    receiver address is invalid

recieverInvalid3
    [Tags]    invalidAdd2
    [Template]    setInvalidReciever
    f    -32000    receiver address is invalid

recieverEmpty
    [Tags]    invalidAdd1
    [Template]    setInvalidReciever
    ${Empty}    -32000    receiver address is empty

*** Keywords ***
