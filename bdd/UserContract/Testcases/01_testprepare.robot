*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn
Library           Collections

*** Test Cases ***
testprepare
    Given queryTokenHolder
    And startProduce
    And Unlock token holder succeed
    When Set token holder as a developer
    Then Token holder can query him in developer list

*** Keywords ***
startProduce
    ${port}    set variable    ${8645}
    ${hosts}    Create List
    : FOR    ${n}    IN RANGE    ${nodenum}
    \    Run Keyword If    ${n}==${0}    startNodeProduce    ${host}
    \    Continue For Loop If    ${n}==${0}
    \    ${newport}=    Evaluate    ${port}+10*(${n}+1)
    \    ${url}=    Catenate    SEPARATOR=    http://${ip}:    ${newport}    /
    \    Append To List    ${hosts}    ${url}
    \    startNodeProduce    ${url}
    Set Global Variable    ${juryHosts}    ${hosts}

Set token holder as a developer
    # step1 unlock account
    unlockAccount    ${tokenHolder}
    # depoist some PTN as a developer
    ${args}=    Create List    DeveloperPayToDepositContract
    ${params}=    Create List    ${tokenHolder}    ${depositContractAddr}    ${1}    ${1}    ${args}
    ${respJson}=    sendRpcPost    ${host}    contract_depositContractInvoke    ${params}    DepositInvoke
    Dictionary Should Contain Key    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${respJson}    result
    wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

Token holder can query him in developer list
    ${args}=    Create List    IsInDeveloperList    ${tokenHolder}
    ${params}=    Create List    ${args}
    ${respJson}=    sendRpcPost    ${host}    contract_depositContractQuery    ${params}    QueryDeposit
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Should Be Equal    ${result}    true
