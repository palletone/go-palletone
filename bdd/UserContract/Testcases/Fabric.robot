*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn
Library           Collections

*** Variables ***
${fabricTxID}     ${EMPTY}

*** Test Cases ***
InstallContractpayTpl
    [Tags]    fabric
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/fabsample    example
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

DeployContract
    [Tags]    fabric
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

SetFabricChaincode
    [Tags]    fabric
    Given Unlock token holder succeed
    ${reqId} =    Invoke contract to set fabricChaincode
    sleep    10
    GetTxByReqId    ${reqId}

Payout
    [Tags]    fabric
    Given Unlock token holder succeed
    When User transfer PTN to contract
    And Wait for transaction being packaged
    And Query contract balance
    ${newAddr}    ${reqId}=    And Invoke contract to payout PTN to user2
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    Then Query user2 balance    ${newAddr}

Invoke
    [Tags]    fabric
    Given Unlock token holder succeed
    When Invoke contract to invoke fabricChaincode
    And Wait for transaction being packaged


*** Keywords ***
User transfer PTN to contract
    transferPtnTo    ${gContractId}    10000
    Wait for transaction being packaged

Query contract balance
    ${amount}=    getBalance    ${gContractId}    PTN
    Should Be Equal    ${amount}    10000
    Log    ${amount}

Invoke contract to payout PTN to user2
    # create account
    ${newAddr}=    newAccount
    Log    ${newAddr}
    ${args}=    Create List    payoutPTNByTxID    ${fabricTxID}    ${newAddr}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    request_id
    [Return]    ${newAddr}    ${reqId}

GetTxByReqId
    [Arguments]    ${reqId}
    ${params}    Create List    ${reqId}
    ${respJson}=    sendRpcPost    ${host}    dag_getTxByReqId    ${params}    GetTxByReqId
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    log    ${respJson}
    ${errCode}=    Evaluate    re.findall('\"error_code\":(\\d*)', '${result}')    re
    ${errMsg}=    Evaluate    re.findall('\"error_message\":\"([^"]*)\"', '${result}')    re
    log    ${errCode}=
    log    ${errMsg}=

Invoke contract to set fabricChaincode
    ${args}=    Create List    setFabChaincodeID    chaincode_example02
    ${respJson}=    invokeContract    ${tokenHolder}    ${gContractId}    0    1    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    request_id
    [Return]    ${reqId}

Invoke contract to invoke fabricChaincode
    ${args}=    Create List    invokeFabChaincode
    ${respJson}=    invokeContract    ${tokenHolder}    ${gContractId}    0.000002    1    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    request_id
    [Return]    ${reqId}

Query user2 balance
    [Arguments]    ${addr}
    [Documentation]    根据fabric链码转账交易，A账户发给B账户100，对应PTN个数0.000001
    ${amount}=    getBalance    ${addr}    PTN
    Should Be Equal    ${amount}    0.000001
    Log    ${amount}
