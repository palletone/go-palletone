*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn
Library           Collections
Library           String

*** Test Cases ***
InstallTestshimucTpl
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/testshimuc1    testshimuc
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

DeployTestshimuc
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

AddState
    Given Unlock token holder succeed
    ${reqId}=    When User put state    testPutState    state1    state1
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    ${reqId}=    When User put state    testPutState    state2    state2
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    # -------- put global state should be error ----------
    ${reqId}=    And User put state    testPutGlobalState    gState1    gState1
    ${errCode}    ${errMsg}=    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${false}
    Should Be Equal    ${errMsg}    Chaincode Error:Only system contract can call this function.
    # -------- query contract state --------------
    Then User query state    testGetState    state1    state1    str    ${null}
    Then User query state    testGetState    state2    state2    str    ${null}
    And User query state    testGetContractState    state1    state1    str    ${gContractId}
    And User query state    testGetContractState    state2    state2    str    ${gContractId}
    And User query state    testGetContractState    gState1    ${EMPTY}    str    ${gContractId}
    And User query state    testGetGlobalState    gState1    ${EMPTY}    str    ${null}
    ${allState}=    And Create Dictionary    state1    state1    state2    state2
    And User query state    testGetStateByPrefix    state    ${allState}    dict    ${null}
    ${allState}=    And Create Dictionary    paystate0    paystate0    state1    state1    state2
    ...    state2
    And User query state    testGetContractAllState    ${null}    ${allState}    dict    ${null}

DelState
    Given Unlock token holder succeed
    ${reqId}=    When User delete state    testDelState    state1
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    ${reqId}=    And User delete state    testDelState    state2
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    Then User query state    testGetState    state1    ${EMPTY}    str    ${null}
    And User query state    testGetGlobalState    state2    ${EMPTY}    str    ${null}
    And User query state    testGetContractState    state1    ${EMPTY}    str    ${gContractId}
    And User query state    testGetContractState    state2    ${EMPTY}    str    ${gContractId}
    And User query state    testGetStateByPrefix    state    ${null}    dict    ${null}
    ${allState}=    And Create Dictionary    paystate0    paystate0
    And User query state    testGetContractAllState    ${null}    ${allState}    dict    ${null}

HandleToken
    Given Unlock token holder succeed
    # -> define token
    ${reqId}=    When User define token    my token    YY    1    100000
    ${errCode}    ${errMsg}=    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${false}
    Should Be Equal    ${errMsg}    Chaincode Error:Only system contract can call this function.
    # -> supply token
    ${reqId}=    And User supply token    YY    100000    1
    ${errCode}    ${errMsg}=    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${false}
    Should Be Equal    ${errMsg}    Chaincode Error:Only system contract can call this function.
    # -> payout token
    ${newAddr}=    Then newAccount
    And User transfer PTN to testshimuc
    ${reqId}=    And User pay out token    ${newAddr}    PTN    4500
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    # -> query balance
    Then Query balance by contract    ${tokenHolder}    YY    ${0}
    ${exceptedAmount}=    And getBalance    ${tokenHolder}    PTN
    ${exceptedAmount}=    And Evaluate    ${exceptedAmount}*${100000000}
    And Query balance by contract    ${tokenHolder}    PTN    ${exceptedAmount}
    And Query balance by contract    ${newAddr}    PTN    ${4500}

UseDigitalCertificate
    Sleep    120s
    Given Unlock token holder succeed
    And queryCAHolder
    And queryCACertID
    # -> use cert
    ${reqId}=    Then User use cert to invoke contract
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    ${payload}=    Then Get invoke payload info    ${reqId}
    ${compareBytes}=    And Replace String    ${caCertBytes}    \n    ${EMPTY}
    ${compareBytes}=    And Replace String    ${compareBytes}    -----BEGIN CERTIFICATE-----    ${EMPTY}
    ${compareBytes}=    And Replace String    ${compareBytes}    -----END CERTIFICATE-----    ${EMPTY}
    And Should Be Equal    ${payload}    ${compareBytes}

TestSendRecvJury
    Given Unlock token holder succeed
    ${reqId}=    When Test send and recv jury by contract
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

TestSetEvent
    Given Unlock token holder succeed
    ${reqId}=    When Test set_event by contract
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

Get Invoke Info
    Given Unlock token holder succeed
    ${args}=    And Create List    arg1    arg2
    ${newAddr}=    newAccount
    ${reqId}=    When User get invoke info    ${args}    ${newAddr}
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    ${payload}=    Get invoke payload info    ${reqId}
    Then Check all invoke info    ${payload}    ${args}    testGetInvokeInfo    ${reqId}    ${newAddr}

Stop testshimuc contract
    Given Unlock token holder succeed
    ${reqId}=    Then stopContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

*** Keywords ***
User transfer PTN to testshimuc
    transferPtnTo    ${gContractId}    10000
    Wait for transaction being packaged

Test set_event by contract
    ${args}=    Create List    testGetInvokeInfo
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Test send and recv jury by contract
    ${args}=    Create List    testSetEvent
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User use cert to invoke contract
    ${args}=    Create List    testUseCert
    ${respJson}=    invokeContract    ${caCertHolder}    ${caCertHolder}    100    100    ${gContractId}
    ...    ${args}    ${caCertID}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User get invoke info
    [Arguments]    ${args}    ${newAddr}
    ${newArgs}=    Create List    testGetInvokeInfo
    ${newArgs}=    Combine Lists    ${newArgs}    ${args}
    ${respJson}=    invokeContract    ${tokenHolder}    ${newAddr}    100    100    ${gContractId}
    ...    ${newArgs}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Check all invoke info
    [Arguments]    ${payload}    ${args}    ${exceptedFuncName}    ${reqId}    ${newAddr}
    # => GetArgs
    Dictionary Should Contain Key    ${payload}    GetArgs
    ${GetArgs} =    Get From Dictionary    ${payload}    GetArgs
    List Should Contain Sub List    ${GetArgs}    ${args}
    # => GetStringArgs
    Dictionary Should Contain Key    ${payload}    GetStringArgs
    ${GetStringArgs} =    Get From Dictionary    ${payload}    GetStringArgs
    List Should Contain Sub List    ${GetStringArgs}    ${args}
    # => GetFunctionAndParameters
    Dictionary Should Contain Key    ${payload}    GetFunctionAndParameters
    ${GetFunctionAndParameters} =    Get From Dictionary    ${payload}    GetFunctionAndParameters
    ${funcName}=    Get From Dictionary    ${GetFunctionAndParameters}    functionName
    ${parameters}=    Get From Dictionary    ${GetFunctionAndParameters}    parameters
    Should Be Equal    ${exceptedFuncName}    ${funcName}
    List Should Contain Sub List    ${parameters}    ${args}
    # => GetArgsSlice
    Dictionary Should Contain Key    ${payload}    GetArgsSlice
    ${GetArgsSlice} =    Get From Dictionary    ${payload}    GetArgsSlice
    ${str}=    Evaluate    "".join(${GetArgsSlice})
    ${comp}=    Create List    ${exceptedFuncName}
    ${comp}=    Combine Lists    ${comp}    ${args}
    ${compStr}=    Evaluate    "".join(${comp})
    Should Be Equal    ${str}    ${compStr}
    # => GetTxID
    Dictionary Should Contain Key    ${payload}    GetTxID
    ${GetTxID}=    Get From Dictionary    ${payload}    GetTxID
    ${exceptTxId}=    catenate    SEPARATOR=    0x    ${reqId}
    Should Be Equal    ${GetTxID}    ${exceptTxId}
    # => GetChannelID
    Dictionary Should Contain Key    ${payload}    GetChannelID
    ${GetChannelID} =    Get From Dictionary    ${payload}    GetChannelID
    Should Be Equal    ${GetChannelID}    palletone
    # => GetTxTimestamp
    Dictionary Should Contain Key    ${payload}    GetTxTimestamp
    ${GetTxTimestamp} =    Get From Dictionary    ${payload}    GetTxTimestamp
    # => GetInvokeAddress
    Dictionary Should Contain Key    ${payload}    GetInvokeAddress
    ${GetInvokeAddress} =    Get From Dictionary    ${payload}    GetInvokeAddress
    Should Be Equal    ${GetInvokeAddress}    ${tokenHolder}
    # => GetInvokeTokens
    Dictionary Should Contain Key    ${payload}    GetInvokeTokens
    ${GetInvokeTokens} =    Get From Dictionary    ${payload}    GetInvokeTokens
    # => GetInvokeFees
    Dictionary Should Contain Key    ${payload}    GetInvokeFees
    ${GetInvokeFees} =    Get From Dictionary    ${payload}    GetInvokeFees
    ${amount}=    Get From Dictionary    ${GetInvokeFees}    amount
    ${symbol}=    Get From Dictionary    ${GetInvokeFees}    assetId
    Should Be Equal    ${amount}    ${10000000000}
    Should Be Equal    ${symbol}    PTN
    # => GetContractID
    Dictionary Should Contain Key    ${payload}    GetContractID
    ${GetContractID} =    Get From Dictionary    ${payload}    GetContractID
    Should Be Equal    ${GetContractID}    ${gContractId}
    # => GetInvokeParameters
    Dictionary Should Contain Key    ${payload}    GetInvokeParameters
    ${GetInvokeParameters} =    Get From Dictionary    ${payload}    GetInvokeParameters
    ${funcName}=    Get From Dictionary    ${GetInvokeParameters}    funcName
    Should Be Equal    ${exceptedFuncName}    ${funcName}
    ${invokeAddress}=    Get From Dictionary    ${GetInvokeParameters}    invokeAddress
    Should Be Equal    ${invokeAddress}    ${tokenHolder}
    ${invokeFees}=    Get From Dictionary    ${GetInvokeParameters}    invokeFees
    Dictionaries Should Be Equal    ${GetInvokeFees}    ${invokeFees}
    ${invokeParams}=    Get From Dictionary    ${GetInvokeParameters}    invokeParams
    List Should Contain Sub List    ${invokeParams}    ${args}
    ${invokeTokens}=    Get From Dictionary    ${GetInvokeParameters}    invokeTokens

User define token
    [Arguments]    ${name}    ${symbol}    ${decimal}    ${amount}
    ${args}=    Create List    testDefineToken    ${name}    ${symbol}    ${decimal}    ${amount}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User supply token
    [Arguments]    ${symbol}    ${amount}    ${decimals}
    ${args}=    Create List    testSupplyToken    ${symbol}    ${amount}    ${decimals}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User pay out token
    [Arguments]    ${to}    ${symbol}    ${amount}
    ${args}=    Create List    testPayOutToken    ${to}    ${symbol}    ${amount}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Query balance by contract
    [Arguments]    ${addr}    ${symbol}    ${exceptedAmount}
    ${args}=    Create List    testGetTokenBalance    ${addr}    ${symbol}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    ${result}=    To Json    ${result}
    ${len}=    Get Length    ${result}
    ${amount}=    Run Keyword If    ${len}==${0}    Set Variable    ${0}
    ...    ELSE    Evaluate    ${result}[${0}]["amount"]
    Should Be Equal    ${amount}    ${exceptedAmount}

User put state
    [Arguments]    ${method}    ${key}    ${value}
    ${args}=    Create List    ${method}    ${key}    ${value}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User delete state
    [Arguments]    ${method}    ${key}
    ${args}=    Create List    ${method}    ${key}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User query state
    [Arguments]    ${getmethod}    ${name}    ${exceptedResult}    ${resType}    ${contractId}
    ${args}=    Run Keyword If    '${contractId}'=='${null}'    Create List    ${getmethod}    ${name}
    ...    ELSE    Create List    ${getmethod}    ${contractId}    ${name}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Run Keyword If    '${resType}'=='dict'    Compare Dict    ${result}    ${exceptedResult}
    ...    ELSE IF    '${resType}'=='str'    Should Be Equal    '${result}'    '${exceptedResult}'
    ...    ELSE    Fail    Result type is not supported now.

Compare Dict
    [Arguments]    ${result}    ${exceptedResult}
    ${resDict}=    To Json    ${result}
    ${len}=    Get Length    ${resDict}
    Run Keyword If    ${len}==0 and ${exceptedResult}==${null}    Pass Execution    Result is the expected one
    ...    ELSE IF    ${len}==0 and ${exceptedResult}!=${null}    Fail    Result is not the expected one
