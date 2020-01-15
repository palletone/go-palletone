*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          pubVariables.robot
Resource          pubFuncs.robot
Resource          setups.robot
Library           BuiltIn

*** Variables ***
${Alice}            ${EMPTY}
${Bob}              ${EMPTY}
${AliceToken}       ${EMPTY}
${BobToken}         ${EMPTY}
${foundation}       ${EMPTY}

*** Test Cases ***
exchangemaker
    unlockAccount    ${foundation}
    transferPtn    ${foundation}    ${Alice}    2000
    log    ${Alice}
    log    ${AliceToken}
    transferPtn    ${foundation}    ${Bob}    2000
    sleep    1
    unlockAccount    ${Alice}
    ${AliceTokenID}    Alice issues her personal token, amount is 100000, decimal is 1 succeed    ${Alice}    ${AliceToken}
    sleep    5
    unlockAccount    ${Bob}
    ${BobTokenID}    Bob issues her personal token, amount is 100000, decimal is 1 succeed    ${Bob}    ${BobToken}
    ${onebalance}=    getBalance    ${Alice}    ${AliceTokenID}
    log    ${onebalance}
    ${twobalance}=    getBalance    ${Bob}    ${BobTokenID}
    log    ${twobalance}
    unlockAccount    ${Alice}
    maker    ${Alice}    ${AliceTokenID}    100    ${BobTokenID}    2000
    sleep    5
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn}=    Get From Dictionary    ${reJson[0]}    ExchangeSn
    log    ${exchsn}
    taker    ${Bob}    ${BobTokenID}    2000    ${exchsn}
    sleep    5
    ${afteronebalance}=    getBalance    ${Alice}    ${BobTokenID}
    log    ${afteronebalance}
    ${makeramount}=    Set Variable If    ${afteronebalance}==2000    0    0
    log    ${makeramount}
    ${aftertwobalance}=    getBalance    ${Bob}    ${AliceTokenID}
    log    ${aftertwobalance}
    ${takeramount}=    Set Variable If    ${aftertwobalance}==100    0    0
    log    ${takeramount}
    maker    ${Alice}    ${AliceTokenID}    200    ${BobTokenID}    4000
    sleep    5
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn2}=    Get From Dictionary    ${reJson[0]}    ExchangeSn
    log    ${exchsn2}
    taker    ${Bob}    ${BobTokenID}    2000    ${exchsn2}
    sleep    5
    ${afteronebalance2}=    getBalance    ${Alice}    ${BobTokenID}
    log    ${afteronebalance2}
    ${makeramount}=    Set Variable If    ${afteronebalance2}==4000    0    0
    log    ${makeramount}
    ${aftertwobalance}=    getBalance    ${Bob}    ${AliceTokenID}
    log    ${aftertwobalance}
    ${takeramount}=    Set Variable If    ${aftertwobalance}==200    0    0
    log    ${takeramount}
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn3}=    Get From Dictionary    ${reJson[0]}    ExchangeSn
    log    ${exchsn3}
    cancel    ${Alice}    ${AliceTokenID}    ${exchsn3}
    sleep    5
    ${respJson}    addrexchangequery    ${Alice}
    log    ${respJson}
    run keyword if    ''    in ${respJson}
    ${allmatchquery}    allmatchquery
    log    ${allmatchquery}
    ${matchqueryrespJson}    matchquery    ${Alice}
    log    ${matchqueryrespJson}
    ${respJson}    historyexchangequery
    log    ${respJson}

    maker    ${Alice}    ${AliceTokenID}    300    ${BobTokenID}    9000
    sleep    5
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn4}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn4}

    taker    ${Bob}    ${BobTokenID}    8000    ${exchsn4}
    sleep    5
    ${afteronebalance4}=    getBalance    ${Alice}    ${BobTokenID}
    log    ${afteronebalance4}
    ${makeramount}=    Set Variable If    ${afteronebalance4}=="12000"    true    false     
    log    ${makeramount}
    ${aftertwobalance4}=    getBalance    ${Bob}    ${AliceTokenID}
    log    ${aftertwobalance4}
    ${takeramount}=    Set Variable If    ${aftertwobalance4}=="466.6"    true    false
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
