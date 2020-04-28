*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          pubVariables.robot
Resource          pubFuncs.robot
Library           BuiltIn

*** Variables ***
${Alice}            ${EMPTY}
${Bob}              ${EMPTY}
${AliceToken}       ${EMPTY}
${BobToken}         ${EMPTY}
${foundation}       ${EMPTY}

*** Test Cases ***
addfeetx
    [Documentation]    通过Alice创建token=>创建token交易=>附加手续费交易并签名=》广播交易，校验积分付款是否可用。
    #解锁创世单元并付给Alice和Bob 2000 PTN
    unlockAccount    ${foundation}
    transferPtn    ${foundation}    ${Alice}    2000
    log    ${Alice}
    log    ${AliceToken}
    transferPtn    ${foundation}    ${Bob}    2000
    sleep    1
    #Alice 发行自己的Token
    unlockAccount    ${Alice}
    ${AliceTokenID}=    Alice issues her personal token, amount is 100000, decimal is 1 succeed    ${Alice}    ${AliceToken}
    Set Global Variable    ${AAAliceTokenID}    ${AliceTokenID}
    sleep    5

    #Alice创建不含手续费的交易
    unlockAccount    ${Alice}
    ${rawtx}=    Alice create tx withoutfee    ${AAAliceTokenID}    ${Alice}    ${Bob}    ${amount}    ${pwd}
    log    ${rawtx["hex"]}
    #Alice 将手续费附加到交易中并签名
    unlockAccount    ${Bob}
    ${signedtx}=    Fee and signtx    ${rawtx["hex"]}    ${Bob}    ${fee}    extra    ${pwd}  
    log    ${signedtx}
    log    ${signedtx["result"]}

    #获取签名结果
    ${complete}=    Get From Dictionary    ${signedtx["result"]}    complete
    Should Be Equal    ${complete}    ${true}
    ${signedhex}=    Get From Dictionary    ${signedtx["result"]}    hex
    Wait for transaction being packaged
    #将签名结果广播并展示交易Hash
    ${params}=    Create List    ${signedhex}   
    ${res}=    sendRpcPost    ${sendRawTransaction}    ${params}    sendRawTransaction
    log    ${res}

*** Keywords ***
getBalance
    [Arguments]    ${address}    ${assetId}
    ${two}    Create List    ${address}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${two}
    log    ${result}
    ${len}    Get Length    ${result}
    ${amount}    Set Variable If    ${len}>1     ${result["${assetId}"]}     ${result["PTN"]}
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
