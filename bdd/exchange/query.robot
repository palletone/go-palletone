*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          pubVariables.robot
Resource          pubFuncs.robot
Resource          setups.robot
Library           BuiltIn

*** Variables ***
${addr}           ${EMPTY}
${twotoken}       ${EMPTY}

*** Test Cases ***
addrexchangequery
    log    ${addr}
    unlockAccount    ${addr}
    ${AliceTokenID}=    Alice issues her personal token, amount is 1000, decimal is 1 succeed        ${addr}
    log    ${addr}
    log    ${AliceTokenID}

    ${balance}=    getBalance    ${addr}    ${AliceTokenID}
    log     ${balance}

    ${args}=    Create List    getActiveOrdersByMaker    ${addr}
    ${params}=    Create List    ${exchangeContractAddr}    ${args}    ${1}
    ${addr}    sendRpcPost    ${queryMethod}    ${params}    exchangeQuery
    log    ${respJson}
    log    ${res}
    Dictionary Should Contain Key    ${respJson}    result
    ${res}    Get From Dictionary    ${respJson}    result
    ${reJson}    To Json    ${res}
    ${len}    Get Length    ${reJson}
    ${exchsn}=    Get From Dictionary    ${reJson[0]}    ExchangeSn
    log    ${exchsn}
    ${afteronebalance}=    getBalance    ${addr}    ${twotoken}
    

    ${respJson}    exchangequery
    log    ${respJson}
    ${respJson}    addrexchangequery    ${addr}
    log    ${respJson}
    run keyword if    ''    in ${respJson}
    ${respJson}    addrexchangequery    ${addr}
    log    ${respJson}
    run keyword if    ''    in ${respJson}
    ${allmatchquery}    allmatchquery
    log    ${allmatchquery}
    ${matchqueryrespJson}    matchquery    ${addr}
    log    ${matchqueryrespJson}
    ${respJson}    historyexchangequery
    log    ${respJson}

*** Keywords ***
getBalance
    [Arguments]    ${address}    ${assetId}
    ${two}    Create List    ${address}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${two}
    log    ${result}
    ${len}    Get Length    ${result}
    log    ${len}
    ${amount}    Set Variable If    ${len}==0    0    ${result["${assetId}"]}
    log    ${amount}
    [Return]    ${amount}

getAllBalance
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${getBalanceMethod}    ${params}    getBalance
    Dictionary Should Contain Key    ${respJson}    result
    #Dictionary Should Contain Key    ${respJson["result"]}    ${AliceToken}
    ${result}=    Get From Dictionary    ${respJson}    result
    [Return]    ${result}
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
    [Arguments]    ${addr}
    log    ${addr}
    issueToken    ${addr}    ${AliceToken}    1000    1    addr's
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
