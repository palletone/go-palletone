*** Settings ***
Force Tags        invalidSign
Default Tags      invalidSign
Library           RequestsLibrary
Library           Collections
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         ptn_signRawTransaction

*** Test Cases ***
signTransInvalid1
    [Tags]    invalidSign1
    [Template]    setInvalidSign
    1    -32000    Params decode is invalid
    e    -32000    Params decode is invalid
    ee    -32000    Params decode is invalid
    eee    -32000    Params decode is invalid
    eeee    -32000    Params decode is invalid
    2476876584    -32000    Params decode is invalid

signTransInvalid2
    [Tags]    invalidSign2
    [Template]    setInvalidSign
    @    -32000    Params is invalid
    FDEW    -32000    Params is invalid
    fd#fg    -32000    Params is invalid
    %    -32000    Params is invalid

signTransInvalid3
    [Tags]    invalidSign2
    [Template]    setInvalidSign
    ${Empty}    -32000    Params is empty

*** Keywords ***
