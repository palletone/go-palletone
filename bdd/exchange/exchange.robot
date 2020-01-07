*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          pubVariables.robot
Resource          pubFuncs.robot
Resource          setups.robot
Library           BuiltIn

*** Variables ***
${one}            ${EMPTY}
${two}            ${EMPTY}
${onetoken}       ${EMPTY}
${twotoken}       ${EMPTY}

*** Test Cases ***
exchangemaker
    log    ${one}
    log    ${onetoken}
    ${onebalance}=    getBalance    ${one}    ${onetoken}
    log    ${onebalance}
    ${twobalance}=    getBalance    ${two}    ${twotoken}
    log    ${twobalance}
    #Given Alice issues her personal token, amount is 1000, decimal is 1 succeed    ${one}    ${onetoken}
    #Given Bob issues her #personal token, amount is 1000, decimal is 1 succeed    ${two}    ${twotoken}
    unlockAccount    ${one}
    maker    ${one}    ${onetoken}    100    ${twotoken}    2000
    sleep    3
    exchangequery
    ${respJson}    addrexchangequery    ${one}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn}

    taker    ${two}    ${twotoken}    2000    ${exchsn}
    sleep    10
    ${afteronebalance}=    getBalance    ${one}    ${twotoken}
    log    ${afteronebalance}
    ${makeramount}=    Set Variable If    ${afteronebalance}==2000    0    0
    log    ${makeramount}
    ${aftertwobalance}=    getBalance    ${two}    ${onetoken}
    log    ${aftertwobalance}
    ${takeramount}=    Set Variable If    ${aftertwobalance}==100    0    0
    log    ${takeramount}
*** Keywords ***
getBalance
    [Arguments]    ${address}    ${assetId}
    ${two}    Create List    ${address}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${two}
    log    ${result}
    ${len}    Get Length    ${result}
    ${amount}    Set Variable If    ${len}==0    0    ${result["${assetId}"]}
    [Return]    ${amount}
post
    [Arguments]    ${method}    ${alias}    ${params}
    ${header}    Create Dictionary    Content-Type=application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    http://127.0.0.1:8545
    ${resp}    Post Request    ${alias}    http://127.0.0.1:8545    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    Dictionary Should Contain Key    ${respJson}    result
    ${res}    Get From Dictionary    ${respJson}    result
    [Return]    ${res}

Alice issues her personal token, amount is 1000, decimal is 1 succeed
    unlockAccount    ${Alice}
    log    ${Alice}
    issueToken    ${Alice}    ${AliceToken}    1000    1    Alice's token
    Wait for transaction being packaged
    ${balance}=    getAllBalance    ${Alice}
    ${tokenIDs}=    Get Dictionary Keys    ${balance}
    FOR    ${id}    IN    @{tokenIDs}

Bob issues her personal token, amount is 1000, decimal is 1 succeed
    unlockAccount    ${Bob}
    issueToken    ${Bob}    ${BobToken}    1000    1    Bob's token
    Wait for transaction being packaged
    ${balance}=    getAllBalance    ${Bob}
    ${tokenIDs}=    Get Dictionary Keys    ${balance}
    FOR    ${id}    IN    @{tokenIDs}
    [Return]    ${res}
