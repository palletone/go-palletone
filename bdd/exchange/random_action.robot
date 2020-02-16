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
    [Documentation]    通过Alice和Bob的随机挂单吃单查询等操作，判断功能是否符合预期
    #解锁创世单元地址并给Alice和BOb分别转账2000PTN
    unlockAccount    ${foundation}
    transferPtn    ${foundation}    ${Alice}    2000
    log    ${Alice}
    log    ${AAAliceTokenID}
    transferPtn    ${foundation}    ${Bob}    2000
    #查询并展示Alice有多少AliceToken，并判断是否等于99500
    ${alicealicebalance}=    getBalance    ${Alice}    ${AAAliceTokenID}
    log    ${alicealicebalance}
    Should Be Equal    ${alicealicebalance}    99500

    #查询并展示Bob有多少BobToken，并判断是否等于88000
    ${bobbobbalance}=    getBalance    ${Bob}    ${BBBobTokenID}
    log    ${bobbobbalance}
    Should Be Equal    ${bobbobbalance}    88000

    #查询并展示Alice有多少BobToken，并判断是否等于12000
    ${alicebobbalance}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance}
    Should Be Equal    ${alicebobbalance}    12000
    #查询并展示Bob有多少AliceToken，并判断是否等于466.6
    ${bobalicebalance}=    getBalance    ${Bob}    ${AAAliceTokenID}
    log    ${bobalicebalance}
    Should Be Equal    ${bobalicebalance}    466.6

    #Bob 吃一个不存在的单
    ${taker_resp}=    taker    ${Bob}    ${BBBobTokenID}    2000    0x123456567
    log    ${taker_resp}
    sleep    5

    #查询并展示Alice有多少BobToken，并判断是否等于12000，且和吃单前相同
    ${alicebobbalance2}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance2}
    Should Be Equal    ${alicebobbalance2}    12000
    Should Be Equal    ${alicebobbalance}    ${alicebobbalance2}


    #解锁Alice 并挂单100 AliceToken 换取 2000 BobToken
    unlockAccount    ${Alice}
    maker    ${Alice}    ${AAAliceTokenID}    100    ${BBBobTokenID}    2000
    sleep    5

    #查询所有挂单并查询,展示所有Alice挂单
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    # here is error  find a wrong ExchangeSn
    ${exchsn}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn}
    
    #从所有挂单中找出匹配汇率的挂单并展示
    :FOR    ${id}    IN    @{reJson}
    \    log    ${id}
    \    ${SaleAmount}=    Get From Dictionary    ${id}    SaleAmount
    \    ${WantAmount}=    Get From Dictionary    ${id}    WantAmount
    \    ${exchsn}=    Get From Dictionary    ${id}    ExchangeSn
    \    ${rate}=    Evaluate    ${WantAmount}/${SaleAmount}
    \    ${old_rate}=    Set Variable    ${2000/100}
    \    run keyword if    '${old_rate}'=='${rate}'    exit for loop

    log    ${exchsn}

    #Bob根据查询结果进行吃单操作
    taker    ${Bob}    ${BBBobTokenID}    4000    ${exchsn}
    sleep    5
    #查询交易后Alice 名下的BobToken，并判断是否等于14000
    ${alicebobbalance3}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance3}
    Should Be Equal    ${alicebobbalance3}    14000
    #查询交易后Bob 名下的AliceToken，并判断是否等于566.6
    ${bobalicebalance3}=    getBalance    ${Bob}    ${AAAliceTokenID}
    log    ${bobalicebalance3}
    Should Be Equal    ${bobalicebalance3}    566.6
    
    #Alice 并挂单300 AliceToken 换取 4000 BobToken
    maker    ${Alice}    ${AAAliceTokenID}    200    ${BBBobTokenID}    4000
    sleep    5

    #查询所有挂单并查询,展示所有Alice挂单
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn2}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn2}

    #从所有挂单中找出匹配汇率的挂单并展示
    :FOR    ${id}    IN    @{reJson}
    \    log    ${id}
    \    ${SaleAmount}=    Get From Dictionary    ${id}    SaleAmount
    \    ${WantAmount}=    Get From Dictionary    ${id}    WantAmount
    \    ${exchsn2}=    Get From Dictionary    ${id}    ExchangeSn
    \    ${rate}=    Evaluate    ${WantAmount}/${SaleAmount}
    \    ${old_rate}=    Set Variable    ${2000/100}
    \    run keyword if    '${old_rate}'=='${rate}'    exit for loop

    log    ${exchsn2}
    #Bob根据查询结果进行吃单操作
    taker    ${Bob}    ${BBBobTokenID}    2000    ${exchsn2}
    sleep    5

    #查询交易后Alice 名下的BobToken，并判断是否等于16000
    ${alicebobbalance4}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance4}
    Should Be Equal    ${alicebobbalance4}    16000

    #查询交易后Bob 名下的AliceToken，并判断是否等于666.6
    ${bobalicebalance4}=    getBalance    ${Bob}    ${AAAliceTokenID}
    log    ${bobalicebalance4}
    Should Be Equal    ${bobalicebalance4}    666.6

    #查询Alice 名下的所有挂单
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn3}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn3}
     #从所有挂单中找出匹配汇率的挂单并展示
    :FOR    ${id}    IN    @{reJson}
    \    log    ${id}
    \    ${SaleAmount}=    Get From Dictionary    ${id}    SaleAmount
    \    ${WantAmount}=    Get From Dictionary    ${id}    WantAmount
    \    ${exchsn3}=    Get From Dictionary    ${id}    ExchangeSn
    \    ${rate}=    Evaluate    ${WantAmount}/${SaleAmount}
    \    ${old_rate}=    Set Variable    ${2000/100}
    \    run keyword if    '${old_rate}'=='${rate}'    exit for loop

    log    ${exchsn3}
    #查询交易后Alice 名下的AliceToken，并展示
    ${alicealicebalance}=    getBalance    ${Alice}    ${AAAliceTokenID}
    log    ${alicealicebalance}
    #取消查询到的订单
    cancel    ${Alice}    ${AAAliceTokenID}    ${exchsn3}
    sleep    5
    #查询交易后Alice 名下的AliceToken，并展示
    ${alicealicebalance}=    getBalance    ${Alice}    ${AAAliceTokenID}
    log    ${alicealicebalance}

    #查询Alice 名下所有订单
    ${respJson}    addrexchangequery    ${Alice}
    log    ${respJson}
    run keyword if    ''    in ${respJson}
    #查询所有成交订单并展示
    ${allmatchquery}    allmatchquery
    log    ${allmatchquery}
    #查询Alice所有成交订单并展示
    ${matchqueryrespJson}    matchquery    ${Alice}
    log    ${matchqueryrespJson}
    #查询所有历史成交订单并展示
    ${respJson}    historyexchangequery
    log    ${respJson}

    #Bob吃单，付出2000 BobToken
    taker    ${Bob}    ${BBBobTokenID}    2000    ${exchsn3}
    sleep    5
    #获取Alice名下所有BobToken，并判断是否等于16000
    ${alicebobbalance5}=    getBalance    ${Alice}    ${BBBobTokenID}
    log    ${alicebobbalance5}
    Should Be Equal    ${alicebobbalance5}    16000
    #获取Bob名下所有AliceToken，并判断是否等于666.6
    ${bobalicebalance5}=    getBalance    ${Bob}    ${AAAliceTokenID}
    log    ${bobalicebalance5}
    Should Be Equal    ${bobalicebalance5}    666.6

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

