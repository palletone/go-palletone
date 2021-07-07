*** Settings ***
Resource          pubVariables.robot
Library           Collections

*** Keywords ***
genInvoketxParams
    [Arguments]    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}    ${args}
    ${params}=    Create List    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}
    ...    ${args}
    [Return]    ${params}

newAccount
    ${params}=    Create List    ${pwd}
    ${respJson}=    sendRpcPost    personal_newAccount    ${params}    newAccount
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson["result"]}

transferPtnTo
    [Arguments]    ${to}
    ${params}=    Create List    ${tokenHolder}    ${to}    ${amount}    ${fee}    ${null}
    ...    ${pwd}
    ${respJson}=    sendRpcPost    ${transferPTNMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result

transferTokenTo
    [Arguments]    ${tokenID}    ${from}    ${to}    ${amount}    ${fee}
    ${params}=    Create List    ${tokenID}    ${from}    ${to}    ${amount}    ${fee}
    ...    ${null}    ${pwd}
    ${respJson}=    sendRpcPost    ${transferTokenMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result

getBalance
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${getBalanceMethod}    ${params}    getBalance
    log    ${respJson}
    Dictionary Should Contain Key    ${respJson}    result
    Dictionary Should Contain Key    ${respJson["result"]}    ${gasToken}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${amount}=    Get From Dictionary    ${result}    ${gasToken}
    [Return]    ${amount}

getAllBalance
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${getBalanceMethod}    ${params}    getBalance
    Dictionary Should Contain Key    ${respJson}    result
    Dictionary Should Contain Key    ${respJson["result"]}    ${gasToken}
    ${result}=    Get From Dictionary    ${respJson}    result
    [Return]    ${result}

unlockAccount
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}    ${pwd}    ${0}
    ${respJson}=    sendRpcPost    ${unlockAccountMethod}    ${params}    unlockAccount
    Dictionary Should Contain Key    ${respJson}    result
    ${res}=    Get From Dictionary    ${respJson}    result
    Should Be Equal    ${res}    ${true}
    [Return]    ${respJson}

issueToken
    [Arguments]    ${addr}    ${name}    ${amount}    ${decimal}    ${des}
    ${args}=    Create List    createToken    ${des}    ${name}    ${decimal}    ${amount}
    ...    ${addr}
    ${params}=    genInvoketxParams    ${addr}    ${addr}    100    1    ${prc720ContractAddr}
    ...    ${args}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    issueToken
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson}

supplyToken
    [Arguments]    ${addr}    ${tokenID}    ${amount}
    ${args}=    Create List    supplyToken    ${tokenID}    ${amount}    ${addr}
    ${params}=    genInvoketxParams    ${addr}    ${addr}    100    1    ${prc720ContractAddr}
    ...    ${args}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    supplyToken
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson}

sendRpcPost
    [Arguments]    ${method}    ${params}    ${alias}
    ${header}=    Create Dictionary    Content-Type    application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    ${host}
    ${resp}    POST On Session    ${alias}    ${host}    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    [Return]    ${respJson}

Wait for transaction being packaged
    Log    Wait for transaction being packaged
    Sleep    6
