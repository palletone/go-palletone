*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn
Library           Collections

*** Test Cases ***
InstallTestshimucTpl
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/testshimuc    testshimuc
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
    ${reqId}=    When User define token    my token    YY    1    100000
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    And Query balance by contract    ${tokenHolder}    ${assetId}    10000
    ${reqId}=    And User supply token    YY    100000
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    And Query balance by contract    ${tokenHolder}    YY    20000
    ${newAddr}=    Then newAccount
    ${reqId}=    And User pay out token    ${newAddr}    YY    4500
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    And Query balance by contract    ${tokenHolder}    YY    19550
    And Query balance by contract    ${newAddr}    YY    450

Get Invoke Info
    Given Unlock token holder succeed
    ${args}=    And Create List    arg1    arg2
    ${reqId}=    When User get invoke info    ${args}
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    ${payload}=    Get invoke payload info    ${reqId}
    Then Check all invoke info    ${payload}    ${args}    testGetInvokeInfo

Stop testshimuc contract
    Given Unlock token holder succeed
    ${reqId}=    Then stopContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

*** Keywords ***
User get invoke info
    [Arguments]    ${args}
    ${newArgs}=    Create List    testGetInvokeInfo
    ${newArgs}=    Combine Lists    ${newArgs}    ${args}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    ...    ${newArgs}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Check all invoke info
    [Arguments]    ${payload}    ${args}    ${exceptedFuncName}    ${reqId}
    # => GetArgs
    Dictionary Should Contain Key    ${payload}    GetArgs
    ${GetArgs} =    Get From Dictioanry    ${payload}    GetArgs
    List Should Contain Sub List    ${GetArgs}    ${args}
    # => GetStringArgs
    Dictionary Should Contain Key    ${payload}    GetStringArgs
    ${GetStringArgs} =    Get From Dictioanry    ${payload}    GetStringArgs
    List Should Contain Sub List    ${GetStringArgs}    ${args}
    # => GetFunctionAndParameters
    Dictionary Should Contain Key    ${payload}    GetFunctionAndParameters
    ${GetFunctionAndParameters} =    Get From Dictioanry    ${payload}    GetFunctionAndParameters
    ${funcName}=    Get From Dictionary    ${GetFunctionAndParameters}    functionName
    ${parameters}=    Get From Dictioanry    ${GetFunctionAndParameters}    parameters
    Shoulb Be Equal    ${exceptedFuncName}    ${funcName}
    List Should Contain Sub List    ${parameters}    ${args}
    # => GetArgsSlice
    Dictionary Should Contain Key    ${payload}    GetArgsSlice
    ${GetArgsSlice} =    Get From Dictioanry    ${payload}    GetArgsSlice
    ${str}=    Evaluate    "".join(${GetArgsSlice})
    ${comp}=    Create List    ${exceptedFuncName}
    ${comp}=    CrCombine Lists    ${comp}    ${args}
    Should Be Equal    ${str}    ${comp}
    # => GetTxID
    Dictionary Should Contain Key    ${payload}    GetTxID
    ${GetTxID}=    Get From Dictioanry    ${payload}    GetTxID
    ${exceptTxId}=    catenate    SEPARATOR=    0x    ${reqId}
    Sould Be Equal    ${GetTxID}    ${exceptTxId}
    # => GetChannelID
    Dictionary Should Contain Key    ${payload}    GetChannelID
    ${GetChannelID} =    Get From Dictioanry    ${payload}    GetChannelID
    Should Be Equal    ${GetChannelID}    palletone
    # => GetTxTimestamp
    Dictionary Should Contain Key    ${payload}    GetTxTimestamp
    ${GetTxTimestamp} =    Get From Dictioanry    ${payload}    GetTxTimestamp
    # => GetInvokeAddress
    Dictionary Should Contain Key    ${payload}    GetInvokeAddress
    ${GetInvokeAddress} =    Get From Dictioanry    ${payload}    GetInvokeAddress
    Should Be Equal    ${GetInvokeAddress}    ${tokenHolder}
    # => GetInvokeTokens
    Dictionary Should Contain Key    ${payload}    GetInvokeTokens
    ${GetInvokeTokens} =    Get From Dictioanry    ${payload}    GetInvokeTokens
    # => GetInvokeFees
    Dictionary Should Contain Key    ${payload}    GetInvokeFees
    ${GetInvokeFees} =    Get From Dictioanry    ${payload}    GetInvokeFees
    ${amount}=    Get From Dictionary    ${GetInvokeFees}    ${amount}
    ${symbol}=    Get From Dictionary    ${GetInvokeFees}    ${assetId}
    Should Be Equal    ${amount}    100000000
    Should Be Equal    ${symbol}    PTN
    # => GetContractID
    Dictionary Should Contain Key    ${payload}    GetContractID
    ${GetContractID} =    Get From Dictioanry    ${payload}    GetContractID
    Should Be Eequal    ${GetContractID}    ${gContractId}
    # => GetInvokeParameters
    Dictionary Should Contain Key    ${payload}    GetInvokeParameters
    ${GetInvokeParameters} =    Get From Dictioanry    ${payload}    GetInvokeParameters
    ${funcName}=    Get From Dictionary    ${GetInvokeParameters}    funcName
    Shoulb Be Equal    ${exceptedFuncName}    ${funcName}
    ${invokeAddress}=    Get From Dictionary    ${GetInvokeParameters}    invokeAddress
    Should Be Equal    ${invokeAddress}    ${tokenHolder}
    ${invokeFees}=    Get From Dictionary    ${GetInvokeParameters}    invokeFees
    Dictionaries Should Be Equal    ${GetInvokeFees}    ${invokeFees}
    ${invokeParams}=    Get From Dictionary    ${GetInvokeParameters}    invokeParams
    List Should Contain Sub List    ${invokeParams}    ${args}
    ${invokeTokens}=    Get From Dictionary    ${GetInvokeParameters}    invokeTokens

User define token
    [Arguments]    ${name}    ${symbole}    ${decimal}    ${amount}
    ${args}=    Create List    testDefineToken    ${name}    ${symbole}    ${decimal}    ${amount}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Query balance by contract
    [Arguments]    ${addr}    ${symbole}    ${exceptedAmount}
    ${args}=    Create List    testGetTokenBalance    ${addr}    ${symbole}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    [Return]    ${reqId}

User put state
    [Arguments]    ${method}    ${key}    ${value}
    ${args}=    Create List    ${method}    ${key}    ${value}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User delete state
    [Arguments]    ${method}    ${key}
    ${args}=    Create List    ${method}    ${key}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
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
