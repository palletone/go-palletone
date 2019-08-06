*** Settings ***
Resource          pubVariables.robot

*** Keywords ***
newAccount
    ${params}=    Create List    1
    ${respJson}=    sendRpcPost    personal_newAccount    ${params}    newAccount
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson["result"]}

transferPTN
    [Arguments]    ${to}
    ${params}=    Create List    ${tokenHolder}    ${to}    ${amount}    ${fee}    ${null}
    ...    ${pwd}
    ${respJson}=    sendRpcPost    ${transerferPTNMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result

getBalance
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${getBalanceMethod}    ${params}    getBalance
    Dictionary Should Contain Key    ${respJson}    result
    Dictionary Should Contain Key    ${respJson["result"]}    PTN
    [Return]    ${respJson["result"]["PTN"]}

unlockAccount
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}    ${pwd}    ${600000000}
    ${respJson}=    sendRpcPost    ${unlockAccountMethod}    ${params}    unlockAccount
    [Return]    ${respJson}

sendRpcPost
    [Arguments]    ${method}    ${params}    ${alias}
    ${header}=    Create Dictionary    Content-Type    application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    http://127.0.0.1:8600
    ${resp}    Post Request    ${alias}    http://127.0.0.1:8600    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    [Return]    ${respJson}

queryToken
    ${args}    Create List    getTokenInfo    ${symbol}
    ${params}    Create List    PCGTta3M4t3yXu8uRgkKvaWd2d8DRijspoq    ${args}    ${0}
    ${resp}    sendRpcPost    ${contractQuery}    ${params}    queryToken
    ${respJson}    To Json   ${resp["result"]}
    ${assetId}    Get From Dictionary    ${respJson}    AssetID
    sleep    6
    [Return]    ${assetId}

queryTokenHolder
    ${args}=    Create List
    ${params}=    Create List
    ${respJson}=    sendRpcPost    ${personalListAccountsMethod}    ${params}    queryTokenHolder
    Dictionary Should Contain Key    ${respJson}    result
    ${accounts}=    Get From Dictionary    ${respJson}    result
    ${firstAddr}=    Get From List    ${accounts}    0
    Set Global Variable    ${tokenHolder}    ${firstAddr}
    log    ${tokenHolder}

wait for transaction being packaged
    Log    wait for transaction being packaged
    Sleep    6
