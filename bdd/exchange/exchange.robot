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
${foundation}     ${EMPTY}

*** Test Cases ***
exchangemaker
    unlockAccount    ${foundation}
    transferPtn    ${foundation}    ${one}    2000
    log    ${one}
    log    ${onetoken}
    transferPtn    ${foundation}    ${two}    2000
    sleep    1
    unlockAccount    ${one}
    ${onetokenId}=    Alice issues her personal token, amount is 100000, decimal is 1 succeed    ${one}    ${onetoken}
    sleep    5
    unlockAccount    ${two}
    ${twotokenId}=    Bob issues her personal token, amount is 100000, decimal is 1 succeed    ${two}    ${twotoken}

    ${onebalance}=    getBalance    ${one}    ${onetokenId}
    log    ${onebalance}
    ${twobalance}=    getBalance    ${two}    ${twotokenId}
    log    ${twobalance}

    unlockAccount    ${one}
    maker    ${one}    ${onetokenId}    100    ${twotokenId}    2000
    sleep    5
    exchangequery
    ${respJson}    addrexchangequery    ${one}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn}

    taker    ${two}    ${twotokenId}    2000    ${exchsn}
    sleep    5
    ${afteronebalance}=    getBalance    ${one}    ${twotokenId}
    log    ${afteronebalance}
    ${makeramount}=    Set Variable If    ${afteronebalance}==2000    0    0
    log    ${makeramount}
    ${aftertwobalance}=    getBalance    ${two}    ${onetokenId}
    log    ${aftertwobalance}
    ${takeramount}=    Set Variable If    ${aftertwobalance}==100    0    0
    log    ${takeramount}

    maker    ${one}    ${onetokenId}    200    ${twotokenId}    4000
    sleep    5
    exchangequery
    ${respJson}    addrexchangequery    ${one}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn2}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn2}

    taker    ${two}    ${twotokenId}    2000    ${exchsn2}
    sleep    5
    ${afteronebalance2}=    getBalance    ${one}    ${twotokenId}
    log    ${afteronebalance2}
    ${makeramount}=    Set Variable If    ${afteronebalance2}==4000    0    0
    log    ${makeramount}
    ${aftertwobalance}=    getBalance    ${two}    ${onetokenId}
    log    ${aftertwobalance}
    ${takeramount}=    Set Variable If    ${aftertwobalance}==200    0    0
    log    ${takeramount}

    ${respJson}    addrexchangequery    ${one}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn3}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn3}
    cancel    ${one}    ${onetokenId}    ${exchsn3}
    sleep    5
    ${respJson}    addrexchangequery    ${one}
    log    ${respJson}
    run keyword if    ''    in ${respJson}
    ${allmatchquery}    allmatchquery
    log    ${allmatchquery}
    ${matchqueryrespJson}    matchquery    ${one}
    log    ${matchqueryrespJson}
    ${respJson}    historyexchangequery
    log    ${respJson}

    maker    ${one}    ${onetokenId}    300    ${twotokenId}    9000
    sleep    5
    exchangequery
    ${respJson}    addrexchangequery    ${one}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn4}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn4}

    taker    ${two}    ${twotokenId}    8000    ${exchsn4}
    sleep    5
    ${afteronebalance4}=    getBalance    ${one}    ${twotokenId}
    log    ${afteronebalance4}
    ${makeramount}=    Set Variable If    ${afteronebalance4}=="12000"    true    false     
    log    ${makeramount}
    ${aftertwobalance4}=    getBalance    ${two}    ${onetokenId}
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

Alice issues her personal token, amount is 100000, decimal is 1 succeed
    [Arguments]    ${addr}    ${AliceToken}
    log    ${addr}
    issueToken    ${addr}    ${AliceToken}    100000    1    addr's
    Wait for transaction being packaged
    ${balance}=    getAllBalance    ${addr}
    log    ${balance}
    ${tokenIDs}=    Get Dictionary Keys    ${balance}
    :FOR    ${id}    IN    @{tokenIDs}
    \    log    ${id[0:3]}
    \    log    ${AliceToken}
    \    Set Global Variable    ${AliceTokenID}    ${id}
    \    run keyword if    '${id[0:3]}'=='${AliceToken}'    exit for loop
    log    ${AliceTokenID}
    [Return]    ${AliceTokenID}

Bob issues her personal token, amount is 100000, decimal is 1 succeed
    [Arguments]    ${addr}    ${BobToken}
    log    ${addr}
    issueToken    ${addr}    ${BobToken}    100000    1    addr's
    Wait for transaction being packaged
    ${balance}=    getAllBalance    ${addr}
    log    ${balance}
    ${tokenIDs}=    Get Dictionary Keys    ${balance}
    :FOR    ${id}    IN    @{tokenIDs}
    \    log    ${id[0:3]}
    \    log    ${BobToken}
    \    Set Global Variable    ${BobTokenID}    ${id}
    \    run keyword if    '${id[0:3]}'=='${BobToken}'    exit for loop
    log    ${BobTokenID}
    [Return]    ${BobTokenID}