*** Settings ***
Resource          ./pubVariables.robot
Library           Collections

*** Keywords ***
genInvoketxParams
    [Arguments]    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}    ${args}    ${certid}
    ${params}=    Create List    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}    ${args}    ${certid}
    [Return]    ${params}

genInvokeExchangeParams
    [Arguments]    ${from}    ${to}    ${assertid}    ${saleamount}    ${fee}    ${exchangeContractAddr}    ${args}
    ${params}=    Create List    ${from}    ${to}    ${assertid}    ${saleamount}    ${fee}    ${exchangeContractAddr}    ${args}
    [Return]    ${params}

newAccount
    ${params}=    Create List    ${pwd}
    ${respJson}=    sendRpcPost    personal_newAccount    ${params}    newAccount
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson["result"]}

transferPtnTo
    [Arguments]    ${to}
    ${params}=    Create List    ${tokenHolder}    ${to}    ${amount}    ${fee}    ${null}    ${pwd}
    ${respJson}=    sendRpcPost    ${transferPTNMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result

transferTokenTo
    [Arguments]    ${tokenID}    ${from}    ${to}    ${amount}    ${fee}
    ${params}=    Create List    ${tokenID}    ${from}    ${to}    ${amount}    ${fee}    ${null}    ${pwd}
    ${respJson}=    sendRpcPost    ${transferTokenMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result

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
    ${params}=    Create List    ${addr}    ${pwd}    ${600000000}
    ${respJson}=    sendRpcPost    ${unlockAccountMethod}    ${params}    unlockAccount
    log    ${respJson}
    #Dictionary Should Contain Key    ${respJson}    result
    #${res}=    Get From Dictionary    ${respJson}    result
    #Should Be Equal    ${res}    True
    Should Be Equal    ${respJson["result"]}    ${true}
    [Return]    ${respJson}

issueToken
    [Arguments]    ${addr}    ${name}    ${amount}    ${decimal}    ${des}
    ${args}=    Create List    createToken    ${des}    ${name}    ${decimal}    ${amount}    ${addr}
    ${params}=    genInvoketxParams    ${addr}    ${addr}    100    1    ${prc720ContractAddr}    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    issueToken
    log    ${respJson}
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson}

supplyToken
    [Arguments]    ${addr}    ${tokenID}    ${amount}
    ${args}=    Create List    supplyToken    ${tokenID}    ${amount}    ${addr}
    ${params}=    genInvoketxParams    ${addr}    ${addr}    100    1    ${prc720ContractAddr}    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    supplyToken
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson}

sendRpcPost
    [Arguments]    ${method}    ${params}    ${alias}
    ${header}=    Create Dictionary    Content-Type    application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    ${host}
    ${resp}    Post Request    ${alias}    ${host}    data=${data}    headers=${header}
    log    ${resp.content}
    ${respJson}    To Json    ${resp.content}
    [Return]    ${respJson}

Wait for transaction being packaged
    Log    Wait for transaction being packaged
    Sleep    6

maker
    [Arguments]    ${makeraddr}    ${saleassert}    ${saleamount}    ${wantassert}    ${wantamount}
    ${args}=    Create List    maker    ${wantassert}    ${wantamount}
    ${params}=    genInvokeExchangeParams    ${makeraddr}    ${exchangeContractAddr}    ${saleassert}    ${saleamount}    ${fee}    ${exchangeContractAddr}    ${args}
    ${resp}=    sendRpcPost    ${exchangeMethod}    ${params}    exchangeToken
    log    ${resp}
    [Return]    ${resp}

exchangequery
    ${args}=    Create List    getActiveOrderList
    ${params}=    Create List    ${exchangeContractAddr}    ${args}    ${1}
    ${resp}=    sendRpcPost    ${queryMethod}    ${params}    exchangeQuery
    [Return]    ${resp}

addrexchangequery
    [Arguments]    ${addr}
    ${args}=    Create List    getActiveOrdersByMaker    ${addr}
    ${params}=    Create List    ${exchangeContractAddr}    ${args}    ${1}
    ${respJson}=    sendRpcPost    ${queryMethod}    ${params}    exchangeQuery
    log    ${respJson}
    log    ${res}
    Dictionary Should Contain Key    ${respJson}    result
    ${res}    Get From Dictionary    ${respJson}    result
    [Return]    ${res}

historyexchangequery
    ${args}=    Create List    getHistoryOrderList
    ${params}=    Create List    ${exchangeContractAddr}    ${args}    ${1}
    ${resp}=    sendRpcPost    ${queryMethod}    ${params}    exchangeQuery
    [Return]    ${resp}

matchquery
    [Arguments]    ${exchange_sn}
    ${args}=    Create List    getOrderMatchList    ${exchange_sn}
    ${params}=    Create List    ${exchangeContractAddr}    ${args}    ${1}
    ${resp}=    sendRpcPost    ${queryMethod}    ${params}    exchangeQuery
    [Return]    ${resp}

allmatchquery
    [Arguments]    ${exchange_sn}
    ${args}=    Create List    getAllMatchList    ${exchange_sn}
    ${params}=    Create List    ${exchangeContractAddr}    ${args}    ${1}
    ${resp}=    sendRpcPost    ${queryMethod}    ${params}    exchangeQuery
    [Return]    ${resp}

taker
    [Arguments]    ${takeraddr}    ${saleassert}    ${saleamount}    ${exchangesn}
    ${args}=    Create List    taker    ${exchangesn}
    ${params}=    genInvokeExchangeParams    ${takeraddr}    ${exchangeContractAddr}    ${saleassert}    ${saleamount}    ${fee}    ${exchangeContractAddr}    ${args}
    ${resp}=    sendRpcPost    ${exchangeMethod}    ${params}    exchangeToken
    log    ${resp}
    [Return]    ${resp}

cancel
    [Arguments]    ${makeraddr}    ${saleassert}    ${exchangesn}
    ${args}=    Create List    cancel    ${exchangesn}
    ${params}=    genInvokeExchangeParams    ${makeraddr}    ${makeraddr}    ${saleassert}    0    ${fee}    ${exchangeContractAddr}    ${args}
    ${resp}=    sendRpcPost    ${exchangeMethod}    ${params}    exchangeCancel
    log    ${resp}
    [Return]    ${resp}
