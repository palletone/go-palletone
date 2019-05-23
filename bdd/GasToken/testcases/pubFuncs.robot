*** Settings ***
Resource          pubVariables.robot

*** Keywords ***
newAccount
    ${params}=    Create List    1
    ${respJson}=    sendRpcPost    personal_newAccount    ${params}    newAccount
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson["result"]}

transferPtnTo
    [Arguments]    ${to}
    ${params}=    Create List    ${tokenHolder}    ${to}    ${amount}    ${fee}    ${null}
    ...    ${pwd}
    ${respJson}=    sendRpcPost    ${transferPTNMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result

getBalance
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${getBalanceMethod}    ${params}    getBalance
    Dictionary Should Contain Key    ${respJson}    result
    Dictionary Should Contain Key    ${respJson["result"]}    ${gasToken}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${amount}=    Get From Dictionary    ${result}    ${gasToken}
    [Return]    ${amount}

unlockAccount
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}    ${pwd}    ${600000000}
    ${respJson}=    sendRpcPost    ${unlockAccountMethod}    ${params}    unlockAccount
    [Return]    ${respJson}

sendRpcPost
    [Arguments]    ${method}    ${params}    ${alias}
    ${header}=    Create Dictionary    Content-Type    application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    http://127.0.0.1:8545
    ${resp}    Post Request    ${alias}    http://127.0.0.1:8545    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    [Return]    ${respJson}

wait for transaction being packaged
    Log    wait for transaction being packaged
    Sleep    10
