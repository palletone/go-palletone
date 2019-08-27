*** Settings ***
Resource          pubVariables.robot
Library           Collections
Library           RequestsLibrary
Library           String
Library           BuiltIn

*** Keywords ***
queryCAHolder
    ${args}=    Create List    getRootCAHoler
    ${params}=    Create List    ${certContractAddr}    ${args}    ${0}
    ${respJson}=    sendRpcPost    ${host}    ${ccqueryMethod}    ${params}    getCAHolder
    Dictionary Should Contain Key    ${respJson}    result
    ${addr}=    Get From Dictionary    ${respJson}    result
    Set Global Variable    ${caCertHolder}    ${addr}
    [Return]    ${addr}

queryCACertID
    ${args}=    Create List    getHolderCertIDs    ${caCertHolder}
    ${params}=    Create List    ${certContractAddr}    ${args}    ${0}
    ${respJson}=    sendRpcPost    ${host}    ${ccqueryMethod}    ${params}    getCAHolder
    ${result}=    Get From Dictionary    ${respJson}    result
    ${info}=    To Json    ${result}
    ${certId}=    Evaluate    ${info}["IntermediateCertIDs"][${0}]["CertID"]
    Should Not Be Empty    ${certId}
    Set Global Variable    ${caCertID}    ${certId}
    [Return]    ${certId}

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
    [Arguments]    ${addr}    ${symbol}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${host}    ${getBalanceMethod}    ${params}    getBalance
    Dictionary Should Contain Key    ${respJson}    result
    Dictionary Should Contain Key    ${respJson["result"]}    ${symbol}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${amount}=    Get From Dictionary    ${result}    ${symbol}
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
    ${respJson}=    sendRpcPost    ${host}    ${ccinvokeMethod}    ${params}    issueToken
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson}

supplyToken
    [Arguments]    ${addr}    ${tokenID}    ${amount}
    ${args}=    Create List    supplyToken    ${tokenID}    ${amount}    ${addr}
    ${params}=    genInvoketxParams    ${addr}    ${addr}    100    1    ${prc720ContractAddr}
    ...    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${host}    ${ccinvokeMethod}    ${params}    supplyToken
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
    ...    ${tplPath}    ${tplVersion}    ${null}    ${null}    go    ${null}
    ${respJson}=    sendRpcPost    ${host}    ${ccinstallMethod}    ${params}    InstallContractTpl
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Dictionary Should Contain Key    ${result}    reqId
    Dictionary Should Contain Key    ${result}    tplId
    [Return]    ${respJson}

deployContract
    [Arguments]    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${tplId}    ${args}
    ${params}=    Create List    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${tplId}
    ...    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${host}    ${ccdeployMethod}    ${params}    DeployContract
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Dictionary Should Contain Key    ${result}    ContractId
    Dictionary Should Contain Key    ${result}    reqId
    [Return]    ${respJson}

invokeContract
    [Arguments]    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${contractId}    ${args}
    ...    ${certId}=${null}
    ${params}=    Create List    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${contractId}
    ...    ${args}    ${certId}    0
    ${respJson}=    sendRpcPost    ${host}    ${ccinvokeMethod}    ${params}    InvokeContract
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Dictionary Should Contain Key    ${result}    ContractId
    Dictionary Should Contain Key    ${result}    reqId
    [Return]    ${respJson}

stopContract
    [Arguments]    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${contractId}
    ${params}=    Create List    ${from}    ${to}    ${ptnAmount}    ${ptnFee}    ${contractId}
    ${respJson}=    sendRpcPost    ${host}    ${ccstopMethod}    ${params}    InvokeContract
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    [Return]    ${result}

queryContract
    [Arguments]    ${contractId}    ${args}
    ${params}=    Create List    ${contractId}    ${args}    ${0}
    Should Not Be Empty    ${juryHosts}
    ${juryHost}=    Get From List    ${juryHosts}    0
    ${respJson}=    sendRpcPost    ${juryHost}    ${ccqueryMethod}    ${params}    QueryContract
    [Return]    ${respJson}

getCurrentUnitHeight
    [Arguments]    ${host}
    # query current unit height
    ${params}=    Create List    PTN
    ${respJson}=    sendRpcPost    ${host}    dag_getFastUnitIndex    ${params}    QueryCurrentUnitHeight
    ${result}=    Get From Dictionary    ${respJson}    result
    ${result}=    To Json    ${result}
    ${stableIndex}=    Get From Dictionary    ${result}    fast_index
    [Return]    ${stableIndex}

sendRpcPost
    [Arguments]    ${host}    ${method}    ${params}    ${alias}
    ${header}=    Create Dictionary    Content-Type    application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    ${host}
    ${resp}    Post Request    ${alias}    ${host}    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    [Return]    ${respJson}

Wait for transaction being packaged
    Log    Wait for transaction being packaged
    Sleep    10s

Unlock token holder succeed
    unlockAccount    ${tokenHolder}
    Log    "unlock ${tokenHolder} succeed"

User installs contract template
    [Arguments]    ${path}    ${name}
    ${respJson}=    installContractTpl    ${tokenHolder}    ${tokenHolder}    100    100    jury06
    ...    ${path}    ${name}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${tplId}=    Get From Dictionary    ${result}    tplId
    Run Keyword If    '${tplId}'=='${EMPTY}'    Fail    "Install Contract Error"
    Set Global Variable    ${gTplId}    ${tplId}
    [Return]    ${reqId}

User deploys contract
    ${args}=    Create List    A    1000
    ${respJson}=    deployContract    ${tokenHolder}    ${tokenHolder}    1000    100    ${gTplId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Run Keyword If    '${contractId}'=='${EMPTY}'    Fail    "Deploy Contract Error"
    Set Global Variable    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User stops contract
    ${args}=    Create List    stop
    ${respJson}=    deployContract    ${tokenHolder}    ${tokenHolder}    1000    100    ${gTplId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Run Keyword If    '${contractId}'=='${EMPTY}'    Fail    "Deploy Contract Error"
    Set Global Variable    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Wait for unit about contract to be confirmed by unit height
    [Arguments]    ${reqId}    ${checkCode}
    Wait for transaction being packaged
    # query the height of unit including tpl install tx
    ${params}=    Create List    ${reqId}
    # jury signature needs time to set agree
    ${waitTimes}=    Set Variable    ${60}
    ${unitHeight}=    Set Variable    999999999999999999999999999999999999
    : FOR    ${t}    IN RANGE    ${waitTimes}
    \    ${respJson}=    sendRpcPost    ${host}    dag_getTxByReqId    ${params}    QueryContractReqStats
    \    ${status}    ${result}=    Run Keyword And Ignore Error    Get From Dictionary    ${respJson}    result
    \    Exit For Loop If    '${status}' == 'PASS'
    \    Run Keyword If    ${waitTimes}-${t}==1    Fail    "It takes too long for jury to signature"
    \    Sleep    5s
    # ------- query error code ------- #
    ${errCode}    ${errMsg}=    Query Error Msg From Response    ${result}
    Run Keyword If    ${checkCode}==${true}    Check response code    ${errCode}    ${errMsg}
    # ------- end of query error code ------- #
    ${result}=    To Json    ${result}
    ${info}=    Get From Dictionary    ${result}    info
    ${unitHeight}=    Get From Dictionary    ${info}    unit_height
    # query current unit height
    ${waitTimes}=    Set Variable    ${8}
    : FOR    ${t}    IN RANGE    ${waitTimes}
    \    ${height}=    getCurrentUnitHeight    ${host}    # query current unit height
    \    Run Keyword If    ${height}-${unitHeight}>3    Exit For Loop
    \    Run Keyword If    ${waitTimes}-${t}==1    Fail    "It takes too long to confirm unit"
    \    Sleep    3s
    [Return]    ${errCode}    ${errMsg}

Check response code
    [Arguments]    ${resCode}    ${resMsg}
    Run Keyword If    ${resCode}!=0    Fail    ${resMsg}

Get invoke payload info
    [Arguments]    ${reqId}
    ${params}=    Create List    ${reqId}
    ${respJson}=    sendRpcPost    ${host}    dag_getTxByReqId    ${params}    QueryContractReqStats
    ${result}=    Get From Dictionary    ${respJson}    result
    ${result}=    To Json    ${result}
    ${info}=    Get From Dictionary    ${result}    info
    ${invokeInfo}=    Get From Dictionary    ${info}    contract_invoke
    ${payload}=    Get From Dictionary    ${invokeInfo}    payload
    ${payload}=    To Json    ${payload}
    [Return]    ${payload}

Query Error Msg From Response
    [Arguments]    ${resStr}
    ${res}=    To Json    ${resStr}
    ${info}=    Get From Dictionary    ${res}    info
    # template
    ${tpl}=    Evaluate    ${info}.get("contract_tpl", {})
    ${type}=    Evaluate    type(${tpl}).__name__
    ${tplLen}=    Run Keyword If    '${type}'=='NoneType'    Set Variable    ${0}
    ...    ELSE    Get Length    ${tpl}
    # deploy
    ${deploy}=    Get From Dictionary    ${info}    contract_deploy
    ${type}=    Evaluate    type(${deploy}).__name__
    ${deployLen}=    Run Keyword If    '${type}'=='NoneType'    Set Variable    ${0}
    ...    ELSE    Get Length    ${deploy}
    # invoke
    ${invoke}=    Get From Dictionary    ${info}    contract_invoke
    ${type}=    Evaluate    type(${invoke}).__name__
    ${invokeLen}=    Run Keyword If    '${type}'=='NoneType'    Set Variable    ${0}
    ...    ELSE    Get Length    ${invoke}
    # stop
    ${stop}=    Get From Dictionary    ${info}    contract_stop
    ${type}=    Evaluate    type(${stop}).__name__
    ${stopLen}=    Run Keyword If    '${type}'=='NoneType'    Set Variable    ${0}
    ...    ELSE    Get Length    ${stop}
    # query error code
    ${errCode}    ${errMsg}=    Run Keyword If    ${tplLen}>0    queryErr    ${tpl}
    ...    ELSE IF    ${deployLen}>0    queryErr    ${deploy}
    ...    ELSE IF    ${invokeLen}>0    queryErr    ${invoke}
    ...    ELSE IF    ${stopLen}>0    queryErr    ${stop}
    ...    ELSE    Set Variable    ${-1}    Get no error info from response
    [Return]    ${errCode}    ${errMsg}

queryErr
    [Arguments]    ${res}
    ${errCode}=    Get From Dictionary    ${res}    error_code
    ${errMsg}=    Get From Dictionary    ${res}    error_message
    [Return]    ${errCode}    ${errMsg}
