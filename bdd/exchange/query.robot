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
    #${args}=    Create List    getActiveOrdersByMaker    ${addr}
    #${params}=    Create List    ${exchangeContractAddr}    ${args}    ${1}
    #${respJson}=    sendRpcPost    ${queryMethod}    ${params}    exchangeQuery
    #log    ${respJson}
    #log    ${res}
    #Dictionary Should Contain Key    ${respJson}    result
    #${res}    Get From Dictionary    ${respJson}    result
    #${reJson}    To Json    ${res}
    #${len}    Get Length    ${reJson}
    #${exchsn}=    Get From Dictionary    ${reJson[0]}    ExchangeSn
    #log    ${exchsn}
    ${afteronebalance}=    getBalance    ${addr}    ${twotoken}
    log    ${afteronebalance}

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