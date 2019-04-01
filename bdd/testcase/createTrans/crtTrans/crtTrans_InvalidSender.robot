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
senderInvalidFormat1
    [Tags]    invalidAdd1
    [Template]    setInvalidSender
    P1    -32000    sender address is invalid

senderInvalidFormat2
    [Tags]    invalidAdd2
    [Template]    setInvalidSender
    P    -32000    sender address is invalid

senderInvalidFormat3
    [Tags]    invalidAdd2
    [Template]    setInvalidSender
    f    -32000    sender address is invalid

senderEmpty
    [Tags]    invalidAdd2
    [Template]    setInvalidSender
    ${Empty}    -32000    sender address is empty

*** Keywords ***
