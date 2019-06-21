*** Settings ***
Resource          pubVariables.robot
Library           Collections

*** Keywords ***
genInvoketxParams
    [Arguments]    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}    ${args}
    ...    ${certid}
    ${params}=    Create List    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}
    ...    ${args}    ${certid}    0
    [Return]    ${params}

newAccount
    ${params}=    Create List    ${pwd}
    ${respJson}=    sendRpcPost    ${host}    personal_newAccount    ${params}    newAccount
    Dictionary Should Contain Key    ${respJson}    result
    ${addr}=    Get From Dictionary    ${respJson}    result
    [Return]    ${addr}

transferPtnTo
    [Arguments]    ${to}    ${num}
    ${params}=    Create List    ${tokenHolder}    ${to}    ${num}    ${fee}    ${null}
    ...    ${pwd}
    ${respJson}=    sendRpcPost    ${host}    ${transferPTNMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result

transferTokenTo
    [Arguments]    ${tokenID}    ${from}    ${to}    ${amount}    ${fee}
    ${params}=    Create List    ${tokenID}    ${from}    ${to}    ${amount}    ${fee}
    ...    ${null}    ${pwd}
    ${respJson}=    sendRpcPost    ${host}    ${transferTokenMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result

getBalance
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${host}    ${getBalanceMethod}    ${params}    getBalance
    Dictionary Should Contain Key    ${respJson}    result
    Dictionary Should Contain Key    ${respJson["result"]}    ${gasToken}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${amount}=    Get From Dictionary    ${result}    ${gasToken}
    [Return]    ${amount}

getAllBalance
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${host}    ${getBalanceMethod}    ${params}    getBalance
    Dictionary Should Contain Key    ${respJson}    result
    Dictionary Should Contain Key    ${respJson["result"]}    ${gasToken}
    ${result}=    Get From Dictionary    ${respJson}    result
    [Return]    ${result}

unlockAccount
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}    ${pwd}    ${600000000}
    ${respJson}=    sendRpcPost    ${host}    ${unlockAccountMethod}    ${params}    unlockAccount
    Dictionary Should Contain Key    ${respJson}    result
    ${res}=    Get From Dictionary    ${respJson}    result
    Should Be Equal    ${res}    ${true}
    [Return]    ${respJson}

issueToken
    [Arguments]    ${addr}    ${name}    ${amount}    ${decimal}    ${des}
    ${args}=    Create List    createToken    ${des}    ${name}    ${decimal}    ${amount}
    ...    ${addr}
    ${params}=    genInvoketxParams    ${addr}    ${addr}    100    1    ${prc720ContractAddr}
    ...    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${host}    ${invokeMethod}    ${params}    issueToken
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson}

supplyToken
    [Arguments]    ${addr}    ${tokenID}    ${amount}
    ${args}=    Create List    supplyToken    ${tokenID}    ${amount}    ${addr}
    ${params}=    genInvoketxParams    ${addr}    ${addr}    100    1    ${prc720ContractAddr}
    ...    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${host}    ${invokeMethod}    ${params}    supplyToken
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson}

startNodeProduce
    [Arguments]    ${host}
    ${respJson}=    sendRpcPost    ${host}    mediator_startProduce    ${null}    startproduce
    ${result}=    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${result}    ${null}

installContractTpl
    [Arguments]    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${tplName}    ${tplPath}
    ...    ${tplVersion}
    ${params}=    Create List    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${tplName}
    ...    ${tplPath}    ${tplVersion}    ${null}
    ${respJson}=    sendRpcPost    ${host}    ${installMethod}    ${params}    InstallContractTpl
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Dictionary Should Contain Key    ${result}    reqId
    Dictionary Should Contain Key    ${result}    tplId
    [Return]    ${respJson}

deployContract
    [Arguments]    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${tplId}    ${args}
    ${params}=    Create List    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${tplId}
    ...    ${args}
    ${respJson}=    sendRpcPost    ${host}    ${deployMethod}    ${params}    DeployContract
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Dictionary Should Contain Key    ${result}    ContractId
    Dictionary Should Contain Key    ${result}    reqId
    [Return]    ${respJson}

invokeContract
    [Arguments]    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${contractId}    ${args}
    ${params}=    Create List    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${contractId}
    ...    ${args}
    ${respJson}=    sendRpcPost    ${host}    ${invokeMethod}    ${params}    InvokeContract
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Dictionary Should Contain Key    ${result}    ContractId
    Dictionary Should Contain Key    ${result}    reqId
    [Return]    ${respJson}

queryContract
    [Arguments]    ${contractId}    ${args}
    ${params}=    Create List    ${contractId}    ${args}
    ${respJson}=    sendRpcPost    ${host}    ${queryMethod}    ${params}    QueryContract
    [Return]    ${respJson}

getCurrentUnitHeight
    [Arguments]    ${host}
    # query current unit height
    ${respJson}=    sendRpcPost    ${host}    admin_nodeInfo    ${null}    QueryCurrentUnitHeight
    ${protocols}=    Get From Dictionary    ${respJson}    protocols
    ${ptnInfo}=    Get From Dictionary    ${protocols}    PTN
    ${height}=    Get From Dictionary    ${ptnInfo}    Index
    [Return]    ${height}

sendRpcPost
    [Arguments]    ${host}    ${method}    ${params}    ${alias}
    ${header}=    Create Dictionary    Content-Type    application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    ${host}
    ${resp}    Post Request    ${alias}    ${host}    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    [Return]    ${respJson}

wait for transaction being packaged
    Log    wait for transaction being packaged
    Sleep    10

Unlock token holder succeed
    unlockAccount    ${tokenHolder}
    Log    "unlock ${tokenHolder} succeed"

Wait for unit abount contract to be confirmed by unit height
    [Arguments]    ${reqId}
    # query the height of unit including tpl install tx
    ${params}=    Create List    ${reqId}
    ${respJson}=    sendRpcPost    ${host}    dag_getTxByReqId    ${params}    QueryContractReqStats
    ${result}=    Get From Dictionary    ${respJson}    result
    ${result}=    To Json    ${result}
    ${info}=    Get From Dictionary    ${result}    info
    ${unitHeight}=    Get From Dictionary    ${info}    unit_height
    ${waitTimes}=    Set Variable    ${8}
    : FOR    ${t}    IN RANGE    ${waitTimes}
    \    ${height}=    getCurrentUnitHeight    ${host}    # query current unit height
    \    Run Keyword If    ${height}-${unitHeight}>3    Exit For Loop
    \    Run Keyword If    ${waitTimes}-${t}==1    Fail    "It takes too slow to confirm unit"
