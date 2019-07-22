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
${method}         wallet_sendRawTransaction

*** Test Cases ***
sendTransInvalid1
    [Tags]    invalidSign1
    [Template]    InvalidSendTrans
    1    -32000    encodedTx decode is invalid
    e    -32000    encodedTx decode is invalid
    ee    -32000    encodedTx decode is invalid
    eee    -32000    encodedTx decode is invalid
    eeee    -32000    encodedTx decode is invalid
    2476876584    -32000    encodedTx decode is invalid

sendTransInvalid2
    [Tags]    invalidSign2
    [Template]    InvalidSendTrans
    FDEW    -32000    encodedTx is invalid
    fd#fg    -32000    encodedTx is invalid
    %    -32000    encodedTx is invalid

sendTransInvalid3
    [Tags]    invalidSign2
    [Template]    InvalidSendTrans
    \    -32000    Params is Empty
