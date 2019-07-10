*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn

*** Test Cases ***
InstallTestshimucTpl
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/testshimuc    testshimuc
    And wait for transaction being packaged
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}

DeployTestshimuc
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    And wait for transaction being packaged
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}

*** Keywords ***
Getxxxx
    [Arguments]     ${method}       ${key}      ${value}
    Given Unlock token holder succeed
    ${reqId} =    When User put status into testshimuc      ${method}    ${key}    ${value}
    And wait for transaction being packaged
    And Wait for unit about contract to be confirmed by unit height    ${reqId}
    Get status from testshimuc    ${method}    ${key}    ${value}

Generate arguments for test
    ${args}=    Create List    aa    bb    cc
    [Return]    ${args}

Check arguments
    [Arguments]    ${args}
    ${origin}=    Create List    aa    bb    cc
    Lists Should Be Equal    ${origin}    ${args}

User put status into testshimuc
    [Arguments]    ${method}    ${key}    ${value}
    ${args}=    Create List    ${method}    ${key}    ${value}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Get status from testshimuc
    [Arguments]    ${getmethod}    ${name}    ${expectedResult}
    ${args}=    Create List    ${getmethod}    ${name}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Should Be Equal    ${result}    ${expectedResult}
