*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          pubVariables.robot
Resource          pubFuncs.robot
Resource          setups.robot
Library           BuiltIn

*** Variables ***
${Alice}            ${EMPTY}
${Bob}            ${EMPTY}
${foundation}     ${EMPTY}

*** Test Cases ***
random_action
    unlockAccount    ${foundation}
    transferPtn    ${foundation}    ${Alice}    2000
    log    ${Alice}
    log    ${AAAliceTokenID}
    transferPtn    ${foundation}    ${Bob}    2000
    ${alicealicebalance}=    getBalance    ${Alice}    ${AAAliceTokenID}
    log    ${alicealicebalance}
    Should Be Equal    ${alicealicebalance}    99500
    ${bobbobbalance}=    getBalance    ${Bob}    ${BBBobTokenID}
    log    ${bobbobbalance}
    Should Be Equal    ${bobbobbalance}    88000


    ${alicebobbalance}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance}
    Should Be Equal    ${alicebobbalance}    12000

    ${bobalicebalance}=    getBalance    ${Bob}    ${AAAliceTokenID}
    log    ${bobalicebalance}
    Should Be Equal    ${bobalicebalance}    466.6

    ${taker_resp}=    taker    ${Bob}    ${BBBobTokenID}    2000    0x123456567
    log    ${taker_resp}
    sleep    5
    ${alicebobbalance2}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance2}
    Should Be Equal    ${alicebobbalance2}    12000
    Should Be Equal    ${alicebobbalance}    ${alicebobbalance2}


    unlockAccount    ${Alice}
    maker    ${Alice}    ${AAAliceTokenID}    100    ${BBBobTokenID}    2000
    sleep    5

    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    # here is error  find a wrong ExchangeSn

    ${exchsn}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn}

    :FOR    ${id}    IN    @{reJson}
    \    log    ${id}
    \    ${SaleAmount}=    Get From Dictionary    ${id}    SaleAmount
    \    ${WantAmount}=    Get From Dictionary    ${id}    WantAmount
    \    ${exchsn}=    Get From Dictionary    ${id}    ExchangeSn
    \    ${rate}=    Evaluate    ${WantAmount}/${SaleAmount}
    \    ${old_rate}=    Set Variable    ${2000/100}
    \    run keyword if    '${old_rate}'=='${rate}'    exit for loop

    log    ${exchsn}

    taker    ${Bob}    ${BBBobTokenID}    4000    ${exchsn}
    sleep    5

    ${alicebobbalance3}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance3}
    Should Be Equal    ${alicebobbalance3}    14000

    ${bobalicebalance3}=    getBalance    ${Bob}    ${AAAliceTokenID}
    log    ${bobalicebalance3}
    Should Be Equal    ${bobalicebalance3}    566.6

    maker    ${Alice}    ${AAAliceTokenID}    200    ${BBBobTokenID}    4000
    sleep    5
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn2}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn2}

    taker    ${Bob}    ${BBBobTokenID}    2000    ${exchsn2}
    sleep    5
    ${alicebobbalance4}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance4}
    Should Be Equal    ${alicebobbalance4}    15000

    ${bobalicebalance4}=    getBalance    ${Bob}    ${AAAliceTokenID}
    log    ${bobalicebalance4}
    Should Be Equal    ${bobalicebalance4}    600

    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn3}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn3}

    ${alicealicebalance}=    getBalance    ${Alice}    ${AAAliceTokenID}
    log    ${alicealicebalance}

    cancel    ${Alice}    ${AAAliceTokenID}    ${exchsn3}
    sleep    5

    ${alicealicebalance}=    getBalance    ${Alice}    ${AAAliceTokenID}
    log    ${alicealicebalance}

    ${respJson}    addrexchangequery    ${Alice}
    log    ${respJson}
    run keyword if    ''    in ${respJson}
    ${allmatchquery}    allmatchquery
    log    ${allmatchquery}
    ${matchqueryrespJson}    matchquery    ${Alice}
    log    ${matchqueryrespJson}
    ${respJson}    historyexchangequery
    log    ${respJson}

    taker    ${Bob}    ${BBBobTokenID}    2000    ${exchsn2}
    sleep    5
    ${alicebobbalance5}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance5}
    Should Be Equal    ${alicebobbalance5}    15000
    ${bobalicebalance5}=    getBalance    ${Bob}    ${AAAliceTokenID}
    log    ${bobalicebalance5}
    Should Be Equal    ${bobalicebalance5}    600

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

Alice issues her personal token, amount is 10000, decimal is 1 succeed
    [Arguments]    ${addr}    ${AliceToken}
    log    ${addr}
    issueToken    ${addr}    ${AliceToken}    10000    1    addr's
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

