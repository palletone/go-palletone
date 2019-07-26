*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot

*** Test Cases ***
Install
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/container    container
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

Deploy
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    sleep    10

Invoke
    Given Unlock token holder succeed
    ${reqId} =    GetValueWithInvokeAddress
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    Given Unlock token holder succeed
    ${reqId} =    OutOfMemory
    #    ${reqId}    ${true}
    sleep    40
    GetTxByReqId    ${reqId}
    Given Unlock token holder succeed
    ${reqId} =    GetValueWithInvokeAddress
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    Given Unlock token holder succeed
    ${reqId} =    DivideByZero
    sleep    40
    GetTxByReqId    ${reqId}
    Given Unlock token holder succeed
    ${reqId} =    GetValueWithInvokeAddress
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    Given Unlock token holder succeed
    ${reqId} =    IndexOutOfRange
    sleep    40
    GetTxByReqId    ${reqId}
    Given Unlock token holder succeed
    ${reqId} =    GetValueWithInvokeAddress
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    Given Unlock token holder succeed
    ${reqId} =    ForLoop
    sleep    40
    GetTxByReqId    ${reqId}
    Given Unlock token holder succeed
    ${reqId} =    GetValueWithInvokeAddress
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

Stop
    Given Unlock token holder succeed
    ${reqId}=    Then stopContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

*** Keywords ***
Invoke
    [Arguments]    @{args}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    ...    @{args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

GetValueWithInvokeAddress
    ${args}    Create List    GetValueWithInvokeAddress
    ${reqId}    Invoke    ${args}
    [Return]    ${reqId}

GetTxByReqId
    [Arguments]    ${reqId}
    ${params}    Create List    ${reqId}
    ${respJson}=    sendRpcPost    ${host}    dag_getTxByReqId    ${params}    GetTxByReqId
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    log    ${respJson}
    ${errCode}=    Evaluate    re.findall('\"error_code\":(\\d*)', '${result}')    re
    ${errMsg}=    Evaluate    re.findall('\"error_message\":\"([^"]*)\"', '${result}')    re
    Should Be Equal As Numbers    ${errCode[0]}    500
    Should Contain    ${errMsg[0]}    error executing chaincode

OutOfMemory
    [Documentation]    // \ 通过申请大内存导致合约触发OOM
    ...    // \ ["OutOfMemory","10000000000"]
    ...    // \ result := "error executing chaincode: failed to execute transaction: fatal error: runtime: out of memory" or "timeout expired while executing transaction"
    ${args}    Create List    OutOfMemory    1000000000
    ${reqId}    Invoke    ${args}
    [Return]    ${reqId}

DivideByZero
    [Documentation]    // \ 除以 0 导致异常
    ...    // \ ["DivideByZero","10","0"]
    ...    // \ result := "error executing chaincode: failed to execute transaction: panic: runtime error: integer divide by zero" or "timeout expired while executing transaction"
    ${args}    Create List    DivideByZero    10    0
    ${reqId}    Invoke    ${args}
    [Return]    ${reqId}

IndexOutOfRange
    [Documentation]    // \ 越界异常
    ...    // \ ["IndexOutOfRange","10","10"]
    ...    // \ result := "error executing chaincode: failed to execute transaction: panic: runtime error: index outc'd of range" or "timeout expired while executing transaction"
    ${args}    Create List    IndexOutOfRange    10    10
    ${reqId}    Invoke    ${args}
    [Return]    ${reqId}

ForLoop
    [Documentation]    // \ 无限for循环
    ...    // \ ["ForLoop"]
    ...    // \ result := "error executing chaincode: failed to execute transaction: timeout expired while executing transaction"
    ${args}    Create List    ForLoop
    ${reqId}    Invoke    ${args}
    [Return]    ${reqId}
