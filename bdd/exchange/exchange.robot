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
    [Documentation]    通过Alice和Bob的挂单吃单查询等操作，判断功能是否符合预期
    #解锁创世单元地址并给Alice和BOb分别转账2000PTN
    unlockAccount    ${foundation}
    transferPtn    ${foundation}    ${Alice}    2000
    log    ${Alice}
    log    ${AliceToken}
    transferPtn    ${foundation}    ${Bob}    2000
    sleep    1

    #解锁Alice账户并发行AliceToken,数量100000，精度1
    unlockAccount    ${Alice}
    ${AliceTokenID}=    Alice issues her personal token, amount is 100000, decimal is 1 succeed    ${Alice}    ${AliceToken}

    #设置全局变量 AAAliceTokenID  为 AliceTokenID
    Set Global Variable    ${AAAliceTokenID}    ${AliceTokenID}
    sleep    5

    #解锁Bob账户并发行BobToken,数量100000，精度1
    unlockAccount    ${Bob}
    ${BobTokenID}    Bob issues her personal token, amount is 100000, decimal is 1 succeed    ${Bob}    ${BobToken}

    #设置全局变量 BBBobTokenID  为 BobTokenID
    Set Global Variable    ${BBBobTokenID}    ${BobTokenID}

    ${onebalance}=    getBalance    ${Alice}    ${AliceTokenID}
    log    ${onebalance}
    ${twobalance}=    getBalance    ${Bob}    ${BobTokenID}
    log    ${twobalance}

    #Alice挂单卖出100个 AliceToken ，换取2000 BobToken
    unlockAccount    ${Alice}
    maker    ${Alice}    ${AliceTokenID}    100    ${BobTokenID}    2000
    sleep    5

    #查询挂单，并展示单号
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn}=    Get From Dictionary    ${reJson[0]}    ExchangeSn
    log    ${exchsn}

    #Bob 吃单，付出2000个BobToken，将获得100 AliceToken
    taker    ${Bob}    ${BobTokenID}    2000    ${exchsn}
    sleep    5
    #获取并展示交易后Alice 所拥有的BobToken 数量
    ${afteronebalance}=    getBalance    ${Alice}    ${BobTokenID}
    log    ${afteronebalance}
    ${makeramount}=    Set Variable If    ${afteronebalance}==2000    0    0
    log    ${makeramount}
    #获取并展示交易后Bob 所拥有的AliceToken 数量
    ${aftertwobalance}=    getBalance    ${Bob}    ${AliceTokenID}
    log    ${aftertwobalance}
    ${takeramount}=    Set Variable If    ${aftertwobalance}==100    0    0
    log    ${takeramount}

    #Alice挂单卖出200个 AliceToken ，换取4000 BobToken
    maker    ${Alice}    ${AliceTokenID}    200    ${BobTokenID}    4000
    sleep    5

    #查询并展示单号
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn2}=    Get From Dictionary    ${reJson[0]}    ExchangeSn
    log    ${exchsn2}

    #Bob 吃单，付出2000个BobToken，将获得100 AliceToken
    taker    ${Bob}    ${BobTokenID}    2000    ${exchsn2}
    sleep    5

    #获取并展示交易后Alice 所拥有的BobToken 数量，判断是否等于4000
    ${afteronebalance2}=    getBalance    ${Alice}    ${BobTokenID}
    log    ${afteronebalance2}
    ${makeramount}=    Set Variable If    ${afteronebalance2}==4000    0    0
    log    ${makeramount}

    #获取并展示交易后Bob 所拥有的AliceToken 数量，判断是否等于200
    ${aftertwobalance}=    getBalance    ${Bob}    ${AliceTokenID}
    log    ${aftertwobalance}
    ${takeramount}=    Set Variable If    ${aftertwobalance}==200    0    0
    log    ${takeramount}

    #查询并展示Alice名下所有的挂单
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn3}=    Get From Dictionary    ${reJson[0]}    ExchangeSn
    log    ${exchsn3}

    #撤销Alice名下所有的挂单
    cancel    ${Alice}    ${AliceTokenID}    ${exchsn3}
    sleep    5

    #查询并展示Alice名下所有的挂单
    ${respJson}    addrexchangequery    ${Alice}
    log    ${respJson}
    run keyword if    ''    in ${respJson}

    #查询并展示所有的成交记录
    ${allmatchquery}    allmatchquery
    log    ${allmatchquery}
    #查询并展示Alice的成交记录
    ${matchqueryrespJson}    matchquery    ${Alice}
    log    ${matchqueryrespJson}
    #查询并展示历史成交记录
    ${respJson}    historyexchangequery
    log    ${respJson}

    #Alice挂单卖出300个 AliceToken ，换取9000 BobToken
    maker    ${Alice}    ${AliceTokenID}    300    ${BobTokenID}    9000
    sleep    5

    #查询所有挂单，查询Alice的所有挂单并展示
    exchangequery
    ${respJson}    addrexchangequery    ${Alice}
    ${reJson}    To Json    ${respJson}
    ${len}    Get Length    ${reJson}
    ${exchsn4}=    Get From Dictionary    ${reJson[0]}    ExchangeSn 
    log    ${exchsn4}
    
    #Bob 吃单，付出8000个BobToken，将获得100 AliceToken
    taker    ${Bob}    ${BobTokenID}    8000    ${exchsn4}
    sleep    5

    #获取并展示交易后Alice 所拥有的BobToken 数量，判断是否等于12000，false表示没有错误
    ${afteronebalance4}=    getBalance    ${Alice}    ${BobTokenID}
    log    ${afteronebalance4}
    ${makeramount}=    Set Variable If    ${afteronebalance4}=="12000"    true    false     
    log    ${makeramount}

    #获取并展示交易后Bob 所拥有的AliceToken 数量，判断是否等于466.6，false表示没有错误
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
    ${amount}    Set Variable   ${result["${assetId}"]}
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
