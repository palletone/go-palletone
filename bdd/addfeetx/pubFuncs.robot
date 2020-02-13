*** Settings ***
Resource          ./pubVariables.robot
Library           Collections

*** Keywords ***
genInvoketxParams
    [Arguments]    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}    ${args}    ${certid}
    ${params}=    Create List    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}    ${args}    ${certid}
    [Return]    ${params}

genIssuetxParams
    [Arguments]    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}    ${args}    ${certid}    ${timeout}
    ${params}=    Create List    ${caCertHolder}    ${caCertHolder}    ${from}    ${to}    ${certContractAddr}    ${args}    ${certid}    ${timeout}
    [Return]    ${params}

newAccount
    ${params}=    Create List    ${pwd}
    ${respJson}=    sendRpcPost    personal_newAccount    ${params}    newAccount
    Dictionary Should Contain Key    ${respJson}    result
    [Return]    ${respJson["result"]}

transferPtn
    [Arguments]    ${from}    ${to}    ${amount}
    ${params}=    Create List    ${from}    ${to}    ${amount}    ${fee}    ${null}    ${pwd}
    ${respJson}=    sendRpcPost    ${transferPTNMethod}    ${params}    transferPTN
    Dictionary Should Contain Key    ${respJson}    result


getAllBalance
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}
    ${respJson}=    sendRpcPost    ${getBalanceMethod}    ${params}    getBalance
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    [Return]    ${result}

unlockAccount
    [Arguments]    ${addr}
    ${params}=    Create List    ${addr}    ${pwd}    ${600000000}
    ${respJson}=    sendRpcPost    ${unlockAccountMethod}    ${params}    unlockAccount
    log    ${respJson}
    Should Be Equal    ${respJson["result"]}    ${true}
    [Return]    ${respJson}

issueToken
    [Arguments]    ${addr}    ${name}    ${amount}    ${decimal}    ${des}
    ${args}=    Create List    createToken    ${des}    ${name}    ${decimal}    ${amount}    ${addr}
    ${params}=    genIssuetxParams    ${addr}    ${addr}    100    1    ${prc720ContractAddr}    ${args}    1    1
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}     issueToken   
    log    ${respJson}
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
    Dictionary Should Not Contain Key    ${respJson}    error
    [Return]    ${respJson}

Wait for transaction being packaged
    Log    Wait for transaction being packaged
    Sleep    6


Alice issues her personal token, amount is 100000, decimal is 1 succeed
    [Arguments]    ${addr}    ${AliceToken}
    log    ${addr}
    issueToken    ${addr}    ${AliceToken}    100000    1    addr's
    Wait for transaction being packaged
    ${ReturnID}    FindTokenId    ${addr}    ${AliceToken} 
    [Return]    ${ReturnID}

Bob issues her personal token, amount is 100000, decimal is 1 succeed
    [Arguments]    ${addr}    ${BobToken}
    log    ${addr}
    issueToken    ${addr}    ${BobToken}    100000    1    addr's
    Wait for transaction being packaged
    ${ReturnID}    FindTokenId    ${addr}    ${BobToken} 
    [Return]    ${ReturnID}

Alice create tx withoutfee
    [Arguments]    ${TokenId}    ${Alice}    ${Bob}    ${amount}    ${extra}    ${pwd}
    ${paras}    Create List    ${TokenId}    ${Alice}    ${Bob}    ${amount}    ${extra}    ${pwd}
    ${res}    post    ${createTxWithOutFee}    ${createTxWithOutFee}    ${paras}
    log    ${res}

FindTokenId
    [Arguments]    ${addr}    ${TokenName}
    ${balance}=    getAllBalance    ${addr}
    log    ${balance}
    log    ${TokenName}
    ${tokenIDs}=    Get Dictionary Keys    ${balance}
    :FOR    ${id}    IN    @{tokenIDs}
    \    log    ${id[0:3]}
    \    log    ${TokenName}
    \    Set Global Variable    ${TokenID}    ${id}
    \    run keyword if    '${id[0:3]}'=='${TokenName}'    exit for loop
    log    ${TokenID}
    [Return]    ${TokenID}
