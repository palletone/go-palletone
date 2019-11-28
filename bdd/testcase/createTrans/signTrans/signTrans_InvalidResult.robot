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
${method}         wallet_signRawTransaction

*** Test Cases ***
signTransInvalid1
    [Tags]    invalidSign1
    [Template]    InvalidSignTrans
    1    ALL    1    -32000    Params decode is invalid
    e    ALL    1    -32000    Params decode is invalid
    ee    ALL    1    -32000    Params decode is invalid
    eee    ALL    1    -32000    Params decode is invalid
    eeee    ALL    1    -32000    Params decode is invalid
    2476876584    ALL    1    -32000    Params decode is invalid

signTransInvalid2
    [Tags]    invalidSign2
    [Template]    InvalidSignTrans
    @    ALL    1    -32000    Params is invalid
    FDEW    ALL    1    -32000    Params is invalid
    fd#fg    ALL    1    -32000    Params is invalid
    %    ALL    1    -32000    Params is invalid

signTransInvalid3
    [Tags]    invalidSign2
    [Template]    InvalidSignTrans
    ${Empty}    ALL    1    -32000    Params is empty

*** Keywords ***
